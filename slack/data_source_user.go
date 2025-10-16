package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/slack-go/slack"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *slack.Client
}

type UserDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Slack user",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The user ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The username",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("email"),
					}...),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The user's email address",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user *slack.User
	var err error

	if !data.Name.IsNull() {
		user, err = d.searchByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find user by name: %s", err))
			return
		}
	} else if !data.Email.IsNull() {
		user, err = d.client.GetUserByEmailContext(ctx, data.Email.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find user by email: %s", err))
			return
		}
	}

	if user == nil {
		resp.Diagnostics.AddError("Not Found", "No user found matching the criteria")
		return
	}

	data.ID = types.StringValue(user.ID)
	data.Name = types.StringValue(user.Name)
	data.Email = types.StringValue(user.Profile.Email)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *UserDataSource) searchByName(ctx context.Context, name string) (*slack.User, error) {
	users, err := d.client.GetUsersContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get workspace users: %s", err)
	}

	var matchingUsers []slack.User
	for _, user := range users {
		if user.Name == name {
			matchingUsers = append(matchingUsers, user)
		}
	}

	if len(matchingUsers) < 1 {
		return nil, fmt.Errorf("no results found for name %s", name)
	}

	if len(matchingUsers) > 1 {
		return nil, fmt.Errorf("multiple results found for name %s", name)
	}

	return &matchingUsers[0], nil
}
