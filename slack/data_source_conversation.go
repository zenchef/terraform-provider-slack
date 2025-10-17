// Package slack provides Terraform resources and data sources for managing Slack workspaces.
package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

var _ datasource.DataSource = &ConversationDataSource{}

// NewConversationDataSource creates a new Slack conversation data source.
func NewConversationDataSource() datasource.DataSource {
	return &ConversationDataSource{}
}

// ConversationDataSource implements the Slack conversation data source.
type ConversationDataSource struct {
	client *slack.Client
}

// ConversationDataSourceModel describes the data source data model.
type ConversationDataSourceModel struct{
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Topic     types.String `tfsdk:"topic"`
	Purpose   types.String `tfsdk:"purpose"`
	Created   types.Int64  `tfsdk:"created"`
	Creator   types.String `tfsdk:"creator"`
	IsPrivate types.Bool   `tfsdk:"is_private"`
}

// Metadata returns the data source type name.
func (d *ConversationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conversation"
}

// Schema defines the schema for the data source.
func (d *ConversationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Slack conversation",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The conversation ID to look up",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The conversation name",
				Computed:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "The conversation topic",
				Computed:            true,
			},
			"purpose": schema.StringAttribute{
				MarkdownDescription: "The conversation purpose",
				Computed:            true,
			},
			"created": schema.Int64Attribute{
				MarkdownDescription: "Timestamp when the conversation was created",
				Computed:            true,
			},
			"creator": schema.StringAttribute{
				MarkdownDescription: "User ID of the conversation creator",
				Computed:            true,
			},
			"is_private": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is private",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ConversationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*slack.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *slack.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ConversationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConversationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := d.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read conversation: %s", err))
		return
	}

	data.Name = types.StringValue(channel.Name)
	data.Topic = types.StringValue(channel.Topic.Value)
	data.Purpose = types.StringValue(channel.Purpose.Value)
	data.Created = types.Int64Value(int64(channel.Created))
	data.Creator = types.StringValue(channel.Creator)
	data.IsPrivate = types.BoolValue(channel.IsPrivate)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
