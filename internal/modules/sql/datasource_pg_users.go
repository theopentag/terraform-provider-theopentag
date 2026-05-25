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

var _ datasource.DataSource = &pgUsersDataSource{}

type pgUsersDataSource struct {
	client *client.Client
}

type pgUsersModel struct {
	ServerName types.String `tfsdk:"server_name"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	Users      types.List   `tfsdk:"users"`
}

var pgUserAttrTypes = map[string]attr.Type{
	"username":             types.StringType,
	"is_superuser":         types.BoolType,
	"can_create_roles":     types.BoolType,
	"can_create_db":        types.BoolType,
	"can_login":            types.BoolType,
	"is_replication_user":  types.BoolType,
	"password_valid_until": types.StringType,
	"member_of_groups":     types.ListType{ElemType: types.StringType},
}

func NewPGUsersDataSource() datasource.DataSource {
	return &pgUsersDataSource{}
}

func (d *pgUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_pg_users"
}

func (d *pgUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Snapshot of PostgreSQL roles on a managed server (refreshed every 60s by the backend).",
		Attributes: map[string]schema.Attribute{
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SQL-managed server.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last snapshot (ISO UTC).",
			},
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of PostgreSQL roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Computed:    true,
							Description: "Role name.",
						},
						"is_superuser": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role is a superuser.",
						},
						"can_create_roles": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role can create other roles.",
						},
						"can_create_db": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role can create databases.",
						},
						"can_login": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role can log in.",
						},
						"is_replication_user": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role has replication privilege.",
						},
						"password_valid_until": schema.StringAttribute{
							Computed:    true,
							Description: "Password expiry: null = no expiry set, 'infinity' = never expires, ISO string = expiry date.",
						},
						"member_of_groups": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of role groups this user is a member of.",
						},
					},
				},
			},
		},
	}
}

func (d *pgUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pgUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pgUsersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.client.GetPGUsers(ctx, config.ServerName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading pg users", err.Error())
		return
	}

	objs := make([]attr.Value, len(r.Users))
	for i, u := range r.Users {
		passwordValidUntil := types.StringNull()
		if u.PasswordValidUntil != nil {
			passwordValidUntil = types.StringValue(*u.PasswordValidUntil)
		}

		groupElems := make([]attr.Value, len(u.MemberOfGroups))
		for j, g := range u.MemberOfGroups {
			groupElems[j] = types.StringValue(g)
		}
		memberOfGroups := types.ListValueMust(types.StringType, groupElems)

		obj, diags := types.ObjectValue(pgUserAttrTypes, map[string]attr.Value{
			"username":             types.StringValue(u.Username),
			"is_superuser":         types.BoolValue(u.IsSuperuser),
			"can_create_roles":     types.BoolValue(u.CanCreateRoles),
			"can_create_db":        types.BoolValue(u.CanCreateDB),
			"can_login":            types.BoolValue(u.CanLogin),
			"is_replication_user":  types.BoolValue(u.IsReplicationUser),
			"password_valid_until": passwordValidUntil,
			"member_of_groups":     memberOfGroups,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: pgUserAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedAt := types.StringNull()
	if r.UpdatedAt != nil {
		updatedAt = types.StringValue(*r.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, pgUsersModel{
		ServerName: config.ServerName,
		UpdatedAt:  updatedAt,
		Users:      list,
	})...)
}
