package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

var _ resource.Resource = &UsergroupResource{}
var _ resource.ResourceWithImportState = &UsergroupResource{}

func NewUsergroupResource() resource.Resource {
	return &UsergroupResource{}
}

type UsergroupResource struct {
	client *slack.Client
}

type UsergroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
	Channels    types.Set    `tfsdk:"channels"`
	Users       types.Set    `tfsdk:"users"`
}

func (r *UsergroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usergroup"
}

func (r *UsergroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Slack usergroup",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The usergroup ID",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the usergroup",
				Required:            true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "The handle/mention name of the usergroup",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the usergroup",
				Optional:            true,
				Computed:            true,
			},
			"channels": schema.SetAttribute{
				MarkdownDescription: "Channel IDs that the usergroup should be associated with",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "User IDs that are members of the usergroup",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *UsergroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*slack.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *slack.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *UsergroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UsergroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get channel IDs
	var channels []string
	if !data.Channels.IsNull() {
		resp.Diagnostics.Append(data.Channels.ElementsAs(ctx, &channels, false)...)
	}

	userGroup := slack.UserGroup{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Handle:      data.Handle.ValueString(),
		Prefs: slack.UserGroupPrefs{
			Channels: channels,
		},
	}

	createdUserGroup, err := r.client.CreateUserGroupContext(ctx, userGroup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create usergroup: %s", err))
		return
	}

	data.ID = types.StringValue(createdUserGroup.ID)

	// Update members if specified
	if !data.Users.IsNull() {
		var users []string
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &users, false)...)
		if len(users) > 0 {
			_, err := r.client.UpdateUserGroupMembersContext(ctx, createdUserGroup.ID, strings.Join(users, ","))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update usergroup members: %s", err))
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UsergroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UsergroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userGroups, err := r.client.GetUserGroupsContext(ctx, slack.GetUserGroupsOptionIncludeUsers(true))
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
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UsergroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UsergroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get channel IDs
	var channels []string
	if !data.Channels.IsNull() {
		resp.Diagnostics.Append(data.Channels.ElementsAs(ctx, &channels, false)...)
	}

	description := data.Description.ValueString()
	updateOptions := []slack.UpdateUserGroupsOption{
		slack.UpdateUserGroupsOptionName(data.Name.ValueString()),
		slack.UpdateUserGroupsOptionHandle(data.Handle.ValueString()),
		slack.UpdateUserGroupsOptionDescription(&description),
		slack.UpdateUserGroupsOptionChannels(channels),
	}

	_, err := r.client.UpdateUserGroupContext(ctx, data.ID.ValueString(), updateOptions...)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update usergroup: %s", err))
		return
	}

	// Update members
	if !data.Users.IsNull() {
		var users []string
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &users, false)...)
		_, err := r.client.UpdateUserGroupMembersContext(ctx, data.ID.ValueString(), strings.Join(users, ","))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update usergroup members: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UsergroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UsergroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DisableUserGroupContext(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable usergroup: %s", err))
		return
	}
}

func (r *UsergroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
