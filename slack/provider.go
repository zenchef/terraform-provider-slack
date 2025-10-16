package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

// Provider returns a *schema.Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SLACK_TOKEN", nil),
				Description: "The Slack token",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"slack_conversation": resourceSlackConversation(),
			"slack_usergroup":    resourceSlackUserGroup(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"slack_conversation": dataSourceConversation(),
			"slack_user":         dataSourceUser(),
			"slack_usergroup":    dataSourceUserGroup(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	token, ok := d.GetOk("token")
	if !ok {
		return nil, diag.Errorf("could not create slack client. Please provide a token.")
	}

	tokenStr := token.(string)
	if err := validateSlackToken(tokenStr); err != nil {
		return nil, diag.FromErr(err)
	}

	slackClient := slack.New(tokenStr)
	return slackClient, diags
}

func validateSlackToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Validate Slack token format
	// Common Slack token prefixes:
	// xoxb- : Bot tokens
	// xoxp- : User tokens
	// xoxa- : App-level tokens (deprecated)
	// xoxe- : Enterprise Grid tokens
	// xoxr- : Refresh tokens
	// xapp- : App tokens
	validPrefixes := []string{"xoxb-", "xoxp-", "xoxa-", "xoxe-", "xoxr-", "xapp-"}

	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if len(token) >= len(prefix) && token[:len(prefix)] == prefix {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("invalid token format. Slack tokens must start with one of: xoxb-, xoxp-, xoxa-, xoxe-, xoxr-, xapp-")
	}

	return nil
}

func schemaSetToSlice(set *schema.Set) []string {
	s := make([]string, len(set.List()))
	for i, v := range set.List() {
		s[i] = v.(string)
	}
	return s
}

func remove(s []string, r string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != r {
			result = append(result, v)
		}
	}
	return result
}
