package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

var _ datasource.DataSource = &UsergroupDataSource{}

// NewUsergroupDataSource creates a new Slack usergroup data source.
func NewUsergroupDataSource() datasource.DataSource {
	return &UsergroupDataSource{}
}

// UsergroupDataSource implements the Slack usergroup data source.
type UsergroupDataSource struct {
	client *slack.Client
}

// UsergroupDataSourceModel describes the data source data model.
type UsergroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	UsergroupID types.String `tfsdk:"usergroup_id"`
	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
	Users       types.Set    `tfsdk:"users"`
	Channels    types.Set    `tfsdk:"channels"`
}

// Metadata returns the data source type name.
func (d *UsergroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usergroup"
}

// Schema defines the schema for the data source.
func (d *UsergroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Slack usergroup. Either `id` or `name` must be specified, but not both.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The usergroup ID to look up",
				Optional:            true,
				Computed:            true,
			},
			"usergroup_id": schema.StringAttribute{
				MarkdownDescription: "The usergroup ID (same as id)",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The usergroup name to look up or the computed name",
				Optional:            true,
				Computed:            true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "The usergroup handle",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The usergroup description",
				Computed:            true,
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "User IDs that are members of the usergroup",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"channels": schema.SetAttribute{
				MarkdownDescription: "Channel IDs that the usergroup is associated with",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *UsergroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsergroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsergroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of id or name is provided
	hasID := !data.ID.IsNull() && !data.ID.IsUnknown()
	hasName := !data.Name.IsNull() && !data.Name.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Invalid combination of arguments",
			"Either 'id' or 'name' must be specified",
		)
		return
	}

	if hasID && hasName {
		resp.Diagnostics.AddError(
			"Invalid combination of arguments",
			"Only one of 'id' or 'name' can be specified, not both",
		)
		return
	}

	userGroups, err := d.client.GetUserGroupsContext(ctx, slack.GetUserGroupsOptionIncludeUsers(true))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read usergroups: %s", err))
		return
	}

	found := false
	for _, ug := range userGroups {
		// Match by ID or name
		matchByID := hasID && ug.ID == data.ID.ValueString()
		matchByName := hasName && ug.Name == data.Name.ValueString()

		if matchByID || matchByName {
			data.ID = types.StringValue(ug.ID)
			data.UsergroupID = types.StringValue(ug.ID)
			data.Name = types.StringValue(ug.Name)
			data.Handle = types.StringValue(ug.Handle)
			data.Description = types.StringValue(ug.Description)

			channelSet, diags := types.SetValueFrom(ctx, types.StringType, ug.Prefs.Channels)
			resp.Diagnostics.Append(diags...)
			data.Channels = channelSet

			userSet, diags := types.SetValueFrom(ctx, types.StringType, ug.Users)
			resp.Diagnostics.Append(diags...)
			data.Users = userSet

			found = true
			break
		}
	}

	if !found {
		identifier := data.ID.ValueString()
		identifierType := "ID"
		if hasName {
			identifier = data.Name.ValueString()
			identifierType = "name"
		}
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("could not find usergroup with %s: %s", identifierType, identifier))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
