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
var _ provider.Provider = &Provider{}

// Provider defines the provider implementation.
type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance testing.
	version string
}

// ProviderModel describes the provider data model.
type ProviderModel struct {
	Token types.String `tfsdk:"token"`
}

// NewFrameworkProvider creates a new Slack provider factory function.
func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

// Metadata returns the provider type name and version.
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "slack"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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

// Configure prepares a Slack API client for data sources and resources.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel

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

// Resources returns the list of resources supported by this provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewConversationResource,
		NewUsergroupResource,
	}
}

// DataSources returns the list of data sources supported by this provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
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
