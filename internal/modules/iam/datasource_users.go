package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &iamUsersDataSource{}

type iamUsersDataSource struct {
	client *client.Client
}

type iamUsersModel struct {
	Users types.List `tfsdk:"users"`
}

var iamUserAttrTypes = map[string]attr.Type{
	"email":      types.StringType,
	"name":       types.StringType,
	"role":       types.StringType,
	"last_login": types.StringType,
	"created_at": types.StringType,
}

func NewUsersDataSource() datasource.DataSource {
	return &iamUsersDataSource{}
}

func (d *iamUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_users"
}

func (d *iamUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all platform users across all modules (admin role required).",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of platform users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "User email address.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "Platform role: admin, user, or viewer.",
						},
						"last_login": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of last login (ISO UTC). Null if never logged in.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp (ISO UTC).",
						},
					},
				},
			},
		},
	}
}

func (d *iamUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *iamUsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ListIAMUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing IAM users", err.Error())
		return
	}

	objs := make([]attr.Value, len(users))
	for i, u := range users {
		lastLogin := types.StringNull()
		if u.LastLogin != nil {
			lastLogin = types.StringValue(*u.LastLogin)
		}
		obj, diags := types.ObjectValue(iamUserAttrTypes, map[string]attr.Value{
			"email":      types.StringValue(u.Email),
			"name":       types.StringValue(u.Name),
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

	list, diags := types.ListValue(types.ObjectType{AttrTypes: iamUserAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, iamUsersModel{Users: list})...)
}
