package slack

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/slack-go/slack"
)

type testUser struct {
	id    string
	name  string
	email string
}

var (
	testUserCreator testUser
	testUser00      testUser
	testUser01      testUser
)

func init() {
	// Load test users from environment variables
	// Set these in your environment or CI:
	// export SLACK_TEST_USER_CREATOR_ID="UXXXXXXXXX"  # Bot user ID (only ID needed)
	// export SLACK_TEST_USER_00_ID="UXXXXXXXXX"
	// export SLACK_TEST_USER_00_NAME="usernsame1"
	// export SLACK_TEST_USER_00_EMAIL="user1@zenchef.com"
	// export SLACK_TEST_USER_01_ID="UXXXXXXXXX"
	// export SLACK_TEST_USER_01_NAME="username2"
	// export SLACK_TEST_USER_01_EMAIL="user2@zenchef.com"
	testUserCreator = testUser{
		id: os.Getenv("SLACK_TEST_USER_CREATOR_ID"),
		// name and email not used in tests
	}

	testUser00 = testUser{
		id:    os.Getenv("SLACK_TEST_USER_00_ID"),
		name:  os.Getenv("SLACK_TEST_USER_00_NAME"),
		email: os.Getenv("SLACK_TEST_USER_00_EMAIL"),
	}

	testUser01 = testUser{
		id:    os.Getenv("SLACK_TEST_USER_01_ID"),
		name:  os.Getenv("SLACK_TEST_USER_01_NAME"),
		email: os.Getenv("SLACK_TEST_USER_01_EMAIL"),
	}
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"slack": providerserver.NewProtocol6WithError(NewFrameworkProvider("test")()),
}

func testAccPreCheck(t *testing.T) {
	requiredEnvVars := []string{
		"SLACK_TOKEN",
		"SLACK_TEST_USER_CREATOR_ID",
		"SLACK_TEST_USER_00_ID",
		"SLACK_TEST_USER_00_NAME",
		"SLACK_TEST_USER_00_EMAIL",
		"SLACK_TEST_USER_01_ID",
		"SLACK_TEST_USER_01_NAME",
		"SLACK_TEST_USER_01_EMAIL",
	}

	for _, envVar := range requiredEnvVars {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

// testAccProvider is a global test provider instance for backward compatibility
// Note: In Plugin Framework, providers don't have a Meta() method like SDK v2
// Use getTestSlackClient() instead for testing
var testAccProvider provider.Provider

func init() {
	testAccProvider = NewFrameworkProvider("test")()
}

// getTestSlackClient returns a Slack client for testing purposes
// This replaces the old testAccProvider.Meta() pattern from SDK v2
func getTestSlackClient() *slack.Client {
	token := os.Getenv("SLACK_TOKEN")
	return slack.New(token)
}

// archiveConversationWithContext archives a conversation/channel
func archiveConversationWithContext(ctx context.Context, client *slack.Client, channelID string) error {
	err := client.ArchiveConversationContext(ctx, channelID)
	if err != nil && err.Error() != errAlreadyArchived && err.Error() != errChannelNotFound {
		return err
	}
	return nil
}

// findUserGroupByID finds a usergroup by its ID
func findUserGroupByID(ctx context.Context, id string, includeUsers bool, clientInterface interface{}) (*slack.UserGroup, error) {
	client, ok := clientInterface.(*slack.Client)
	if !ok {
		return nil, fmt.Errorf("expected *slack.Client, got %T", clientInterface)
	}

	userGroups, err := client.GetUserGroupsContext(ctx, slack.GetUserGroupsOptionIncludeUsers(includeUsers))
	if err != nil {
		return nil, fmt.Errorf("unable to get usergroups: %s", err)
	}

	for _, ug := range userGroups {
		if ug.ID == id {
			return &ug, nil
		}
	}

	return nil, fmt.Errorf("usergroup %s not found", id)
}
