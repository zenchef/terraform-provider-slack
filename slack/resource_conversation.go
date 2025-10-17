package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

const (
	errChannelNotFound = "channel_not_found"
	errAlreadyArchived = "already_archived"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ConversationResource{}
var _ resource.ResourceWithImportState = &ConversationResource{}

// NewConversationResource creates a new Slack conversation resource.
func NewConversationResource() resource.Resource {
	return &ConversationResource{}
}

// ConversationResource defines the resource implementation
type ConversationResource struct {
	client *slack.Client
}

// ConversationResourceModel describes the resource data model
type ConversationResourceModel struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Topic                          types.String `tfsdk:"topic"`
	Purpose                        types.String `tfsdk:"purpose"`
	PermanentMembers               types.Set    `tfsdk:"permanent_members"`
	Created                        types.Int64  `tfsdk:"created"`
	Creator                        types.String `tfsdk:"creator"`
	IsPrivate                      types.Bool   `tfsdk:"is_private"`
	IsArchived                     types.Bool   `tfsdk:"is_archived"`
	IsShared                       types.Bool   `tfsdk:"is_shared"`
	IsExtShared                    types.Bool   `tfsdk:"is_ext_shared"`
	IsOrgShared                    types.Bool   `tfsdk:"is_org_shared"`
	IsGeneral                      types.Bool   `tfsdk:"is_general"`
	ActionOnDestroy                types.String `tfsdk:"action_on_destroy"`
	ActionOnUpdatePermanentMembers types.String `tfsdk:"action_on_update_permanent_members"`
	AdoptExistingChannel           types.Bool   `tfsdk:"adopt_existing_channel"`
}

func (r *ConversationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conversation"
}

func (r *ConversationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Slack conversation (channel)",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The conversation ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the conversation",
				Required:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "The topic of the conversation",
				Optional:            true,
				Computed:            true,
			},
			"purpose": schema.StringAttribute{
				MarkdownDescription: "The purpose of the conversation",
				Optional:            true,
				Computed:            true,
			},
			"permanent_members": schema.SetAttribute{
				MarkdownDescription: "User IDs who are permanent members of the conversation",
				ElementType:         types.StringType,
				Optional:            true,
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
				Required:            true,
			},
			"is_archived": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is archived",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is shared",
				Computed:            true,
			},
			"is_ext_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is externally shared",
				Computed:            true,
			},
			"is_org_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is org shared",
				Computed:            true,
			},
			"is_general": schema.BoolAttribute{
				MarkdownDescription: "Whether the conversation is the general channel",
				Computed:            true,
			},
			"action_on_destroy": schema.StringAttribute{
				MarkdownDescription: "Action to take when destroying the conversation. Either 'none' or 'archive'. Default is 'archive'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("archive"),
			},
			"action_on_update_permanent_members": schema.StringAttribute{
				MarkdownDescription: "Action to take when updating permanent members. Either 'none' or 'kick'. Default is 'kick'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("kick"),
			},
			"adopt_existing_channel": schema.BoolAttribute{
				MarkdownDescription: "Whether to adopt an existing channel if name is taken",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ConversationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*slack.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *slack.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new Slack conversation resource.
func (r *ConversationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConversationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create conversation using existing logic
	name := data.Name.ValueString()
	isPrivate := data.IsPrivate.ValueBool()

	channel, err := r.client.CreateConversationContext(ctx, slack.CreateConversationParams{
		ChannelName: name,
		IsPrivate:   isPrivate,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create conversation: %s", err))
		return
	}

	// Set basic attributes
	data.ID = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.Creator = types.StringValue(channel.Creator)
	data.Created = types.Int64Value(int64(channel.Created))
	data.IsPrivate = types.BoolValue(channel.IsPrivate)
	data.IsGeneral = types.BoolValue(channel.IsGeneral)
	data.IsShared = types.BoolValue(channel.IsShared)
	data.IsExtShared = types.BoolValue(channel.IsExtShared)
	data.IsOrgShared = types.BoolValue(channel.IsOrgShared)
	data.IsArchived = types.BoolValue(channel.IsArchived)

	// Set optional fields if provided
	if !data.Topic.IsNull() && data.Topic.ValueString() != "" {
		if _, err := r.client.SetTopicOfConversationContext(ctx, channel.ID, data.Topic.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set conversation topic: %s", err))
			return
		}
	}

	if !data.Purpose.IsNull() && data.Purpose.ValueString() != "" {
		if _, err := r.client.SetPurposeOfConversationContext(ctx, channel.ID, data.Purpose.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set conversation purpose: %s", err))
			return
		}
	}

	// Handle permanent members
	if !data.PermanentMembers.IsNull() && len(data.PermanentMembers.Elements()) > 0 {
		var members []string
		resp.Diagnostics.Append(data.PermanentMembers.ElementsAs(ctx, &members, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, userID := range members {
			if _, err := r.client.InviteUsersToConversationContext(ctx, channel.ID, userID); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to invite user %s to conversation: %s", userID, err))
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConversationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConversationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: data.ID.ValueString(),
	})
	if err != nil {
		if err.Error() == errChannelNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read conversation: %s", err))
		return
	}

	// Update state with response
	data.Name = types.StringValue(channel.Name)
	data.Topic = types.StringValue(channel.Topic.Value)
	data.Purpose = types.StringValue(channel.Purpose.Value)
	data.IsArchived = types.BoolValue(channel.IsArchived)
	data.IsShared = types.BoolValue(channel.IsShared)
	data.IsExtShared = types.BoolValue(channel.IsExtShared)
	data.IsOrgShared = types.BoolValue(channel.IsOrgShared)
	data.IsPrivate = types.BoolValue(channel.IsPrivate)
	data.IsGeneral = types.BoolValue(channel.IsGeneral)
	data.Created = types.Int64Value(int64(channel.Created))
	data.Creator = types.StringValue(channel.Creator)

	// Get channel members if permanent_members is set in state
	if !data.PermanentMembers.IsNull() {
		members, _, err := r.client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
			ChannelID: data.ID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get users in conversation: %s", err))
			return
		}

		// Filter out the creator from the members list
		var permanentMembers []string
		creator := channel.Creator
		for _, member := range members {
			if member != creator {
				permanentMembers = append(permanentMembers, member)
			}
		}

		memberSet, diags := types.SetValueFrom(ctx, types.StringType, permanentMembers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.PermanentMembers = memberSet
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates an existing Slack conversation resource.
func (r *ConversationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConversationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	// Update name if changed
	var state ConversationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !data.Name.Equal(state.Name) {
		if _, err := r.client.RenameConversationContext(ctx, id, data.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename conversation: %s", err))
			return
		}
	}

	// Update topic if changed
	if !data.Topic.IsNull() && !data.Topic.Equal(state.Topic) {
		if _, err := r.client.SetTopicOfConversationContext(ctx, id, data.Topic.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set conversation topic: %s", err))
			return
		}
	}

	// Update purpose if changed
	if !data.Purpose.IsNull() && !data.Purpose.Equal(state.Purpose) {
		if _, err := r.client.SetPurposeOfConversationContext(ctx, id, data.Purpose.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set conversation purpose: %s", err))
			return
		}
	}

	// Update archived status if changed
	if !data.IsArchived.Equal(state.IsArchived) {
		if data.IsArchived.ValueBool() {
			if err := r.client.ArchiveConversationContext(ctx, id); err != nil && err.Error() != errAlreadyArchived {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive conversation: %s", err))
				return
			}
		} else {
			if err := r.client.UnArchiveConversationContext(ctx, id); err != nil && err.Error() != "not_archived" {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unarchive conversation: %s", err))
				return
			}
		}
	}

	// Update permanent members if changed
	if !data.PermanentMembers.Equal(state.PermanentMembers) {
		var newMembers, oldMembers []string

		if !data.PermanentMembers.IsNull() {
			resp.Diagnostics.Append(data.PermanentMembers.ElementsAs(ctx, &newMembers, false)...)
		}
		if !state.PermanentMembers.IsNull() {
			resp.Diagnostics.Append(state.PermanentMembers.ElementsAs(ctx, &oldMembers, false)...)
		}

		if resp.Diagnostics.HasError() {
			return
		}

		// Find users to add (in new but not in old)
		for _, userID := range newMembers {
			if !contains(oldMembers, userID) {
				if _, err := r.client.InviteUsersToConversationContext(ctx, id, userID); err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to invite user %s to conversation: %s", userID, err))
					return
				}
			}
		}

		// Find users to remove (in old but not in new)
		action := data.ActionOnUpdatePermanentMembers.ValueString()
		if action == "kick" {
			for _, userID := range oldMembers {
				if !contains(newMembers, userID) {
					if err := r.client.KickUserFromConversationContext(ctx, id, userID); err != nil {
						resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to kick user %s from conversation: %s", userID, err))
						return
					}
				}
			}
		}
	}

	// Refresh state from Slack to ensure it's accurate
	channel, err := r.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read conversation after update: %s", err))
		return
	}

	// Update computed fields with actual values from Slack
	data.Name = types.StringValue(channel.Name)
	data.Topic = types.StringValue(channel.Topic.Value)
	data.Purpose = types.StringValue(channel.Purpose.Value)
	data.IsArchived = types.BoolValue(channel.IsArchived)
	data.IsShared = types.BoolValue(channel.IsShared)
	data.IsExtShared = types.BoolValue(channel.IsExtShared)
	data.IsOrgShared = types.BoolValue(channel.IsOrgShared)

	// Get channel members if permanent_members is set
	if !data.PermanentMembers.IsNull() {
		members, _, err := r.client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
			ChannelID: id,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get users in conversation: %s", err))
			return
		}

		// Filter out the creator from the members list
		var permanentMembers []string
		creator := channel.Creator
		for _, member := range members {
			if member != creator {
				permanentMembers = append(permanentMembers, member)
			}
		}

		memberSet, diags := types.SetValueFrom(ctx, types.StringType, permanentMembers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.PermanentMembers = memberSet
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete removes a Slack conversation resource.
func (r *ConversationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConversationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	action := data.ActionOnDestroy.ValueString()
	if action == "archive" {
		if err := r.client.ArchiveConversationContext(ctx, data.ID.ValueString()); err != nil && err.Error() != errAlreadyArchived && err.Error() != errChannelNotFound {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive conversation: %s", err))
			return
		}
	}
}

// ImportState imports an existing Slack conversation resource.
func (r *ConversationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// contains checks if a string is in a slice
func contains(s []string, e string) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return false
}
