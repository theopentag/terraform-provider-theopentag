package sql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &usersDataSource{}

type usersDataSource struct {
	client *client.Client
}

type usersModel struct {
	Users types.List `tfsdk:"users"`
}

var userAttrTypes = map[string]attr.Type{
	"email":      types.StringType,
	"name":       types.StringType,
	"picture":    types.StringType,
	"role":       types.StringType,
	"last_login": types.StringType,
	"created_at": types.StringType,
}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all users with access to the SQL API (admin role required).",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "User email address (primary key).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name.",
						},
						"picture": schema.StringAttribute{
							Computed:    true,
							Description: "Profile picture URL.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "User role: admin, user, or viewer.",
						},
						"last_login": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of last login (ISO UTC).",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the user was created (ISO UTC).",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing users", err.Error())
		return
	}

	objs := make([]attr.Value, len(users))
	for i, u := range users {
		lastLogin := types.StringNull()
		if u.LastLogin != nil {
			lastLogin = types.StringValue(*u.LastLogin)
		}

		obj, diags := types.ObjectValue(userAttrTypes, map[string]attr.Value{
			"email":      types.StringValue(u.Email),
			"name":       types.StringValue(u.Name),
			"picture":    types.StringValue(u.Picture),
			"role":       types.StringValue(u.Role),
			"last_login": lastLogin,
			"created_at": types.StringValue(u.CreatedAt),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: userAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, usersModel{Users: list})...)
}
