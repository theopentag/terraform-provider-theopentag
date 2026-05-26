package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

type userResource struct {
	client *client.Client
}

type userModel struct {
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	LastLogin types.String `tfsdk:"last_login"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a platform user. Controls access to all modules (SQL, Compute, IAM) via role assignment.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Required:    true,
				Description: "User email address. Immutable — changes force replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Platform role: admin, user, or viewer.",
			},
			"last_login": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of last login (ISO UTC).",
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

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	u, err := r.client.CreateIAMUser(ctx, client.IAMUserCreateRequest{
		Email: plan.Email.ValueString(),
		Name:  name,
		Role:  plan.Role.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating IAM user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, iamUserToModel(u))...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.GetIAMUser(ctx, state.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IAM user", err.Error())
		return
	}
	if u == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, iamUserToModel(u))...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	role := plan.Role.ValueString()
	u, err := r.client.UpdateIAMUser(ctx, plan.Email.ValueString(), client.IAMUserUpdateRequest{
		Name: &name,
		Role: &role,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating IAM user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, iamUserToModel(u))...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIAMUser(ctx, state.Email.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting IAM user", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("email"), req, resp)
}

func iamUserToModel(u *client.IAMUser) userModel {
	m := userModel{
		Email:     types.StringValue(u.Email),
		Name:      types.StringValue(u.Name),
		Role:      types.StringValue(u.Role),
		CreatedAt: types.StringValue(u.CreatedAt),
	}
	if u.LastLogin != nil {
		m.LastLogin = types.StringValue(*u.LastLogin)
	} else {
		m.LastLogin = types.StringNull()
	}
	return m
}
