package slack

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

const (
	userGroupResourceNamePrefix = "test-acc-slack-usergroup-test"
)

func init() {
	resource.AddTestSweepers("slack_usergroup", &resource.Sweeper{
		Name: "slack_useregroup",
		F: func(string) error {
			client, err := sharedSlackClient()
			if err != nil {
				return fmt.Errorf("error getting client: %s", err)
			}
			c := client.(*slack.Client)
			groups, err := c.GetUserGroupsContext(context.Background())
			if err != nil {
				return fmt.Errorf("[ERROR] error getting channels: %s", err)
			}
			var sweeperErrs *multierror.Error
			for _, group := range groups {
				if strings.HasPrefix(group.Name, userGroupResourceNamePrefix) {
					_, err := c.DisableUserGroupContext(context.Background(), group.ID)
					if err != nil {
						if err.Error() != "already_disabled" {
							sweeperErr := fmt.Errorf("disabling usergroup %s during sweep: %s", group.Name, err)
							log.Printf("[ERROR] %s", sweeperErr)
							sweeperErrs = multierror.Append(sweeperErrs, err)
						}
					} else {
						fmt.Printf("[INFO] disabled usergroup %s during sweep\n", group.Name)
					}
				}
			}
			return sweeperErrs.ErrorOrNil()
		},
	})
}

func TestAccSlackUserGroupTest(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set to 1")
		return
	}
	t.Parallel()

	resourceName := "slack_usergroup.test"

	t.Run("update name, description and handle", func(t *testing.T) {
		name := acctest.RandomWithPrefix(userGroupResourceNamePrefix)
		createUserGroup := testAccSlackUserGroup(name)

		updateName := acctest.RandomWithPrefix(userGroupResourceNamePrefix)
		updateUserGroup := testAccSlackUserGroup(updateName)

		testSlackUserGroupUpdate(t, resourceName, createUserGroup, &updateUserGroup)
	})

	t.Run("update users", func(t *testing.T) {
		name := acctest.RandomWithPrefix(userGroupResourceNamePrefix)
		createUserGroup := testAccSlackUserGroupWithUsers(name, []string{}, []string{testUser00.id, testUser01.id})

		updateUserGroup := createUserGroup
		updateUserGroup.Users = []string{testUser00.id}

		testSlackUserGroupUpdate(t, resourceName, createUserGroup, &updateUserGroup)
	})

	t.Run("update channels", func(t *testing.T) {
		channel := createTestConversation(t)

		name := acctest.RandomWithPrefix(userGroupResourceNamePrefix)
		createUserGroup := testAccSlackUserGroupWithUsers(name, []string{}, []string{})

		updateUserGroup := createUserGroup
		updateUserGroup.Prefs = slack.UserGroupPrefs{Channels: []string{channel.ID}}

		testSlackUserGroupUpdate(t, resourceName, createUserGroup, &updateUserGroup)
	})

	t.Run("create without channels attribute", func(t *testing.T) {
		name := acctest.RandomWithPrefix(userGroupResourceNamePrefix)

		// Test creating a usergroup without specifying channels
		// This ensures no "inconsistent result" error when channels is null
		testSlackUserGroupWithoutChannels(t, resourceName, name)
	})
}

func createTestConversation(t *testing.T) *slack.Channel {
	client, err := sharedSlackClient()
	if err != nil {
		require.NoError(t, err, "error getting client: %s", err)
	}

	c := client.(*slack.Client)
	channelName := acctest.RandomWithPrefix(conversationNamePrefix)

	channel, err := c.CreateConversationContext(context.Background(), slack.CreateConversationParams{
		ChannelName: channelName,
		IsPrivate:   false,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = c.ArchiveConversationContext(context.Background(), channel.ID)
	})
	return channel
}

func testSlackUserGroupUpdate(t *testing.T, resourceName string, createChannel slack.UserGroup, updateChannel *slack.UserGroup) {
	steps := []resource.TestStep{
		{
			Config: testAccSlackUserGroupConfig(createChannel),
			Check: resource.ComposeTestCheckFunc(
				testCheckSlackUserGroupAttributes(t, resourceName, createChannel),
				testCheckUserGroupAttrBasic(resourceName, createChannel),
			),
		},
		{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
	}

	if updateChannel != nil {
		steps = append(steps, resource.TestStep{
			Config: testAccSlackUserGroupConfig(*updateChannel),
			Check: resource.ComposeTestCheckFunc(
				testCheckSlackUserGroupAttributes(t, resourceName, *updateChannel),
				testCheckUserGroupAttrBasic(resourceName, *updateChannel),
			),
		},
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName:            resourceName,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy,
		Steps:                    steps,
	})
}

func testCheckUserGroupAttrBasic(resourceName string, channel slack.UserGroup) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", channel.Name),
		resource.TestCheckResourceAttr(resourceName, "description", channel.Description),
		resource.TestCheckResourceAttr(resourceName, "handle", channel.Handle),

		testCheckResourceAttrSlice(resourceName, "users", channel.Users),
		testCheckResourceAttrSlice(resourceName, "channels", channel.Prefs.Channels),
	)
}

func testCheckSlackUserGroupAttributes(t *testing.T, resourceName string, expectedGroup slack.UserGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		primary := rs.Primary
		group, err := findUserGroupByID(context.Background(), primary.ID, true, getTestSlackClient())
		if err != nil {
			return fmt.Errorf("couldn't get conversation info for %s: %s", primary.ID, err)
		}

		require.Equal(t, primary.Attributes["name"], group.Name, "usergroup name does not match")
		require.Equal(t, primary.Attributes["description"], group.Description, "usergroup description does not match")
		require.Equal(t, primary.Attributes["handle"], group.Handle, "usergroup handle does not match")

		usersLength, _ := strconv.Atoi(primary.Attributes["users.#"])
		require.Equal(t, len(expectedGroup.Users), usersLength, "defined users length should match state")
		require.Equal(t, len(expectedGroup.Users), len(group.Users), "defined users length should match users in channel")

		for i := 0; i < usersLength; i++ {
			user := primary.Attributes[fmt.Sprintf("users.%d", i)]
			require.True(t, contains(group.Users, user), "user should be in the group")
			require.True(t, contains(expectedGroup.Users, user), "user in state should be defined in the resource")
		}

		channelsLength, _ := strconv.Atoi(primary.Attributes["channels.#"])
		require.Equal(t, len(expectedGroup.Prefs.Channels), channelsLength, "defined channels length should match state")
		require.Equal(t, len(expectedGroup.Prefs.Channels), len(group.Prefs.Channels), "defined channels length should match users in channel")

		for i := 0; i < channelsLength; i++ {
			channel := primary.Attributes[fmt.Sprintf("channels.%d", i)]
			require.True(t, contains(group.Prefs.Channels, channel), "channel should be in the group")
			require.True(t, contains(expectedGroup.Prefs.Channels, channel), "channel in state should be defined in the resource")
		}

		return nil
	}
}

func testAccSlackUserGroup(name string) slack.UserGroup {
	return testAccSlackUserGroupWithUsers(name, []string{}, []string{})
}

func testAccSlackUserGroupWithUsers(name string, channels, users []string) slack.UserGroup {
	sort.Strings(users)
	// Slack handles can only contain lowercase letters, numbers, and underscores
	// and must be 21 characters or less. Use the last 21 chars to maintain uniqueness.
	handle := strings.ReplaceAll(name, "-", "_")
	if len(handle) > 21 {
		// Take the last 21 characters to keep the unique random suffix
		handle = handle[len(handle)-21:]
	}
	group := slack.UserGroup{
		Name:        name,
		Description: fmt.Sprintf("Description for %s", name),
		Handle:      handle,
		Users:       users,
		Prefs:       slack.UserGroupPrefs{Channels: channels},
	}
	return group
}

func testAccCheckUserGroupDestroy(s *terraform.State) error {
	c := getTestSlackClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "slack_usergroup" {
			continue
		}

		_, err := c.DisableUserGroupContext(context.Background(), rs.Primary.ID)
		if err != nil && err.Error() != "already_disabled" {
			return fmt.Errorf("error disabling usergroup %s: %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccSlackUserGroupConfig(c slack.UserGroup) string {
	var channels, users []string
	for _, channel := range c.Prefs.Channels {
		channels = append(channels, fmt.Sprintf(`"%s"`, channel))
	}
	for _, user := range c.Users {
		users = append(users, fmt.Sprintf(`"%s"`, user))
	}

	return fmt.Sprintf(`
resource slack_usergroup test {
  name        = "%s"
  description = "%s"
  handle      = "%s"
  users       = [%s]
  channels    =  [%s]
}
`, c.Name, c.Description, c.Handle, strings.Join(users, ","), strings.Join(channels, ","))
}

// testSlackUserGroupWithoutChannels tests creating a usergroup without specifying channels attribute
// This ensures the provider doesn't produce "inconsistent result" errors when channels is null
func testSlackUserGroupWithoutChannels(t *testing.T, resourceName string, name string) {
	sort.Strings([]string{testUser00.id})
	handle := strings.ReplaceAll(name, "-", "_")
	if len(handle) > 21 {
		handle = handle[len(handle)-21:]
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName:            resourceName,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlackUserGroupConfigWithoutChannels(name, handle, testUser00.id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Description for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "handle", handle),
					// Verify channels is computed as empty set when not specified
					resource.TestCheckResourceAttr(resourceName, "channels.#", "0"),
					// Verify users is set correctly
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccSlackUserGroupConfigWithoutChannels creates a config without channels attribute
func testAccSlackUserGroupConfigWithoutChannels(name, handle, userID string) string {
	return fmt.Sprintf(`
resource slack_usergroup test {
  name        = "%s"
  description = "Description for %s"
  handle      = "%s"
  users       = ["%s"]
}
`, name, name, handle, userID)
}
