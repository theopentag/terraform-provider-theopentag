package sql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ resource.Resource = &apiKeyResource{}
var _ resource.ResourceWithImportState = &apiKeyResource{}

type apiKeyResource struct {
	client *client.Client
}

type apiKeyModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Key        types.String `tfsdk:"key"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
	CreatedBy  types.String `tfsdk:"created_by"`
	LastUsedAt types.String `tfsdk:"last_used_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SQL API key. The full key value is returned only on creation and cannot be retrieved again.",
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
				Description: "Human-readable name for the API key. Immutable — changes force replacement.",
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
				Description: "Timestamp of last key usage.",
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

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	k, err := r.client.CreateAPIKey(ctx, client.APIKeyCreateRequest{
		Name: plan.Name.ValueString(),
		Role: plan.Role.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating API key", err.Error())
		return
	}

	state := apiKeyModel{
		ID:        types.StringValue(strconv.FormatInt(k.ID, 10)),
		Name:      types.StringValue(k.Name),
		Role:      types.StringValue(k.Role),
		KeyPrefix: types.StringValue(k.KeyPrefix),
		CreatedBy: types.StringValue(k.CreatedBy),
		CreatedAt: types.StringValue(k.CreatedAt),
	}

	if k.Key != nil {
		state.Key = types.StringValue(*k.Key)
	} else {
		state.Key = types.StringNull()
	}
	if k.LastUsedAt != nil {
		state.LastUsedAt = types.StringValue(*k.LastUsedAt)
	} else {
		state.LastUsedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := r.client.ListAPIKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing API keys", err.Error())
		return
	}

	targetID := state.ID.ValueString()
	var found *client.APIKey
	for i := range keys {
		if strconv.FormatInt(keys[i].ID, 10) == targetID {
			found = &keys[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := apiKeyModel{
		ID:        types.StringValue(strconv.FormatInt(found.ID, 10)),
		Name:      types.StringValue(found.Name),
		Role:      types.StringValue(found.Role),
		KeyPrefix: types.StringValue(found.KeyPrefix),
		CreatedBy: types.StringValue(found.CreatedBy),
		CreatedAt: types.StringValue(found.CreatedAt),
		Key:       state.Key,
	}

	if found.LastUsedAt != nil {
		newState.LastUsedAt = types.StringValue(*found.LastUsedAt)
	} else {
		newState.LastUsedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *apiKeyResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// API keys have RequiresReplace on all mutable fields; Update is never called.
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid API key ID", err.Error())
		return
	}

	if err := r.client.DeleteAPIKey(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error deleting API key", err.Error())
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
