package slack

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/slack-go/slack"
)

func TestAccSlackConversationDataSource_basic(t *testing.T) {
	nameByID := acctest.RandomWithPrefix("test-acc-slack-conversation-test")
	resourceNameByID := fmt.Sprintf("slack_conversation.%s", nameByID)
	dataSourceNameByID := fmt.Sprintf("data.slack_conversation.%s", nameByID)
	membersByID := []string{testUser00.id, testUser01.id}
	createChannelByID := testAccSlackConversationWithMembers(nameByID, membersByID)

	nameByName := acctest.RandomWithPrefix("test-acc-slack-conversation-test")
	resourceNameByName := fmt.Sprintf("slack_conversation.%s", nameByName)
	dataSourceNameByName := fmt.Sprintf("data.slack_conversation.%s", nameByName)
	membersByName := []string{testUser00.id, testUser01.id}
	createChannelByName := testAccSlackConversationWithMembers(nameByName, membersByName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckSlackConversationDataSourceConfigNonExistent,
				ExpectError: regexp.MustCompile(`channel_not_found`),
			},
			{
				Config: testAccCheckSlackConversationDataSourceConfig(createChannelByID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackConversationDataSourceID(dataSourceNameByID),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "id", resourceNameByID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "name", resourceNameByID, "name"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "topic", resourceNameByID, "topic"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "purpose", resourceNameByID, "purpose"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "creator", resourceNameByID, "creator"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "created", resourceNameByID, "created"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "is_private", resourceNameByID, "is_private"),
				),
			},
			{
				Config: testAccCheckSlackConversationDataSourceConfig(createChannelByName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackConversationDataSourceID(dataSourceNameByName),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "id", resourceNameByName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "name", resourceNameByName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "topic", resourceNameByName, "topic"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "purpose", resourceNameByName, "purpose"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "creator", resourceNameByName, "creator"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "created", resourceNameByName, "created"),
					resource.TestCheckResourceAttrPair(dataSourceNameByName, "is_private", resourceNameByName, "is_private"),
				),
			},
		},
	})
}

func testAccCheckSlackConversationDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find slack conversation data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("slack conversation data source id not set")
		}
		return nil
	}
}

const (
	testAccCheckSlackConversationDataSourceConfigNonExistent = `
data slack_conversation test {
 id = "non-existent"
}
`
	testAccCheckSlackConversationDataSourceConfigExistent = `
data slack_conversation %s {
  id = slack_conversation.%s.id
}
`
	testAccCheckSlackConversationDataSourceConfigNameExistent = `
data slack_conversation %s {
  name       = slack_conversation.%s.name
  is_private = true
}
`
)

func testAccCheckSlackConversationDataSourceConfig(channel slack.Channel) string {
	return testAccSlackConversationConfig(channel) + fmt.Sprintf(testAccCheckSlackConversationDataSourceConfigExistent, channel.Name, channel.Name)
}

func testAccCheckSlackConversationDataSourceConfigName(channel slack.Channel) string {
	return testAccSlackConversationConfig(channel) + fmt.Sprintf(testAccCheckSlackConversationDataSourceConfigNameExistent, channel.Name, channel.Name)
}
