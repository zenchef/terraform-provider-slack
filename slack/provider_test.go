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
	testUserCreator = testUser{
		id:    "U01D6L97N0M",
		name:  "contact",
		email: "contact@pablovarela.co.uk",
	}

	testUser00 = testUser{
		id:    "U01D31S1GUE",
		name:  "contact_test-user-ter",
		email: "contact+test-user-terraform-provider-slack-00@pablovarela.co.uk",
	}

	testUser01 = testUser{
		id:    "U01DZK10L1W",
		name:  "contact_test-user-206",
		email: "contact+test-user-terraform-provider-slack-01@pablovarela.co.uk",
	}
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"slack": providerserver.NewProtocol6WithError(NewFrameworkProvider("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SLACK_TOKEN"); v == "" {
		t.Fatal("SLACK_TOKEN must be set for acceptance tests")
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

// Helper function to check if a slice contains a string
func contains(s []string, e string) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return false
}

// archiveConversationWithContext archives a conversation/channel
func archiveConversationWithContext(ctx context.Context, client *slack.Client, channelID string) error {
	err := client.ArchiveConversationContext(ctx, channelID)
	if err != nil && err.Error() != "already_archived" && err.Error() != "channel_not_found" {
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
