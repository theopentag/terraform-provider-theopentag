package iam

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &iamAPIKeysDataSource{}

type iamAPIKeysDataSource struct {
	client *client.Client
}

type iamAPIKeysModel struct {
	APIKeys types.List `tfsdk:"api_keys"`
}

var iamAPIKeyAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"role":        types.StringType,
	"scopes":      types.ListType{ElemType: types.StringType},
	"key_prefix":  types.StringType,
	"created_by":  types.StringType,
	"last_used_at": types.StringType,
	"created_at":  types.StringType,
}

func NewAPIKeysDataSource() datasource.DataSource {
	return &iamAPIKeysDataSource{}
}

func (d *iamAPIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_keys"
}

func (d *iamAPIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all platform API keys (admin role required). Full key values are never returned.",
		Attributes: map[string]schema.Attribute{
			"api_keys": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of platform API keys.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "API key ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable name.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "Role: admin, user, or viewer.",
						},
						"scopes": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Module scopes this key has access to. Empty means all modules.",
						},
						"key_prefix": schema.StringAttribute{
							Computed:    true,
							Description: "First 12 characters of the key for display purposes.",
						},
						"created_by": schema.StringAttribute{
							Computed:    true,
							Description: "Identity that created this key.",
						},
						"last_used_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of last key usage (ISO UTC). Null if never used.",
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

func (d *iamAPIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *iamAPIKeysDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	keys, err := d.client.ListIAMAPIKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing IAM API keys", err.Error())
		return
	}

	objs := make([]attr.Value, len(keys))
	for i, k := range keys {
		scopeVals := make([]attr.Value, len(k.Scopes))
		for j, s := range k.Scopes {
			scopeVals[j] = types.StringValue(s)
		}
		scopeList, diags := types.ListValue(types.StringType, scopeVals)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		lastUsedAt := types.StringNull()
		if k.LastUsedAt != nil {
			lastUsedAt = types.StringValue(*k.LastUsedAt)
		}

		obj, diags := types.ObjectValue(iamAPIKeyAttrTypes, map[string]attr.Value{
			"id":          types.StringValue(strconv.FormatInt(k.ID, 10)),
			"name":        types.StringValue(k.Name),
			"role":        types.StringValue(k.Role),
			"scopes":      scopeList,
			"key_prefix":  types.StringValue(k.KeyPrefix),
			"created_by":  types.StringValue(k.CreatedBy),
			"last_used_at": lastUsedAt,
			"created_at":  types.StringValue(k.CreatedAt),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: iamAPIKeyAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, iamAPIKeysModel{APIKeys: list})...)
}
