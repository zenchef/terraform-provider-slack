package slack

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

// Ensure the implementation satisfies the provider.Provider interface
var _ provider.Provider = &SlackProvider{}

// SlackProvider defines the provider implementation
type SlackProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance testing.
	version string
}

// SlackProviderModel describes the provider data model
type SlackProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SlackProvider{
			version: version,
		}
	}
}

func (p *SlackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "slack"
	resp.Version = p.version
}

func (p *SlackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The Slack token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *SlackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SlackProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available
	// Get token from configuration or environment variable
	token := os.Getenv("SLACK_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Slack Token Configuration",
			"While configuring the provider, the Slack token was not found. "+
				"Please set the SLACK_TOKEN environment variable or configure the token in the provider configuration.",
		)
		return
	}

	// Validate token format
	if err := validateSlackToken(token); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Slack Token",
			fmt.Sprintf("The provided Slack token is invalid: %s", err.Error()),
		)
		return
	}

	// Create Slack client
	slackClient := slack.New(token)

	// Make the Slack client available during DataSource and Resource type Configure methods
	resp.DataSourceData = slackClient
	resp.ResourceData = slackClient
}

func (p *SlackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewConversationResource,
		NewUsergroupResource,
	}
}

func (p *SlackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewConversationDataSource,
		NewUserDataSource,
		NewUsergroupDataSource,
	}
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
