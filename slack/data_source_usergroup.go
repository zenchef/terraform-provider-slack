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

func NewUsergroupDataSource() datasource.DataSource {
	return &UsergroupDataSource{}
}

type UsergroupDataSource struct {
	client *slack.Client
}

type UsergroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
}

func (d *UsergroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usergroup"
}

func (d *UsergroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Slack usergroup",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The usergroup ID to look up",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The usergroup name",
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
		},
	}
}

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

	userGroups, err := d.client.GetUserGroupsContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read usergroups: %s", err))
		return
	}

	found := false
	for _, ug := range userGroups {
		if ug.ID == data.ID.ValueString() {
			data.Name = types.StringValue(ug.Name)
			data.Handle = types.StringValue(ug.Handle)
			data.Description = types.StringValue(ug.Description)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Usergroup with ID %s not found", data.ID.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
