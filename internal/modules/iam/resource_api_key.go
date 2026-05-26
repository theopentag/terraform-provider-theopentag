package iam

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ resource.Resource = &iamAPIKeyResource{}
var _ resource.ResourceWithImportState = &iamAPIKeyResource{}

type iamAPIKeyResource struct {
	client *client.Client
}

type iamAPIKeyModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Scopes     types.List   `tfsdk:"scopes"`
	Key        types.String `tfsdk:"key"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
	CreatedBy  types.String `tfsdk:"created_by"`
	LastUsedAt types.String `tfsdk:"last_used_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func NewAPIKeyResource() resource.Resource {
	return &iamAPIKeyResource{}
}

func (r *iamAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key"
}

func (r *iamAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a platform-level API key. Scopes restrict access to specific modules (sql, compute, iam). The full key is returned only on creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "API key ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name. Immutable — changes force replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Role: admin, user, or viewer. Immutable — changes force replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Module scopes this key has access to (e.g. [\"sql\", \"compute\"]). Empty means all modules. Immutable — changes force replacement.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Full API key value (bmk_...). Set only on creation; preserved in state thereafter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "First 12 characters of the key for display purposes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity that created this key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of last key usage (ISO UTC).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (ISO UTC).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *iamAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *iamAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iamAPIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	k, err := r.client.CreateIAMAPIKey(ctx, client.IAMAPIKeyCreateRequest{
		Name:   plan.Name.ValueString(),
		Role:   plan.Role.ValueString(),
		Scopes: scopes,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating IAM API key", err.Error())
		return
	}

	state, diags := iamAPIKeyToModel(ctx, k)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if k.Key != nil {
		state.Key = types.StringValue(*k.Key)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *iamAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid IAM API key ID", err.Error())
		return
	}

	keys, err := r.client.ListIAMAPIKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing IAM API keys", err.Error())
		return
	}

	var found *client.IAMAPIKey
	for i := range keys {
		if keys[i].ID == id {
			found = &keys[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := iamAPIKeyToModel(ctx, found)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	newState.Key = state.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *iamAPIKeyResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable fields have RequiresReplace; Update is never called.
}

func (r *iamAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid IAM API key ID", err.Error())
		return
	}

	if err := r.client.DeleteIAMAPIKey(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error deleting IAM API key", err.Error())
	}
}

func (r *iamAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func iamAPIKeyToModel(ctx context.Context, k *client.IAMAPIKey) (iamAPIKeyModel, diag.Diagnostics) {
	scopeVals := make([]attr.Value, len(k.Scopes))
	for i, s := range k.Scopes {
		scopeVals[i] = types.StringValue(s)
	}
	scopes, diags := types.ListValue(types.StringType, scopeVals)

	m := iamAPIKeyModel{
		ID:        types.StringValue(strconv.FormatInt(k.ID, 10)),
		Name:      types.StringValue(k.Name),
		Role:      types.StringValue(k.Role),
		Scopes:    scopes,
		KeyPrefix: types.StringValue(k.KeyPrefix),
		CreatedBy: types.StringValue(k.CreatedBy),
		CreatedAt: types.StringValue(k.CreatedAt),
		Key:       types.StringNull(),
	}
	if k.LastUsedAt != nil {
		m.LastUsedAt = types.StringValue(*k.LastUsedAt)
	} else {
		m.LastUsedAt = types.StringNull()
	}
	return m, diags
}
