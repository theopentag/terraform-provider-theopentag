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

var _ datasource.DataSource = &serviceDiscoveryDataSource{}

type serviceDiscoveryDataSource struct {
	client *client.Client
}

type serviceDiscoveryModel struct {
	Services  types.List   `tfsdk:"services"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

var serviceAttrTypes = map[string]attr.Type{
	"name":     types.StringType,
	"endpoint": types.StringType,
	"status":   types.StringType,
	"version":  types.StringType,
}

func NewServiceDiscoveryDataSource() datasource.DataSource {
	return &serviceDiscoveryDataSource{}
}

func (d *serviceDiscoveryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_service_discovery"
}

func (d *serviceDiscoveryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads registered service endpoints from the platform's Service Discovery (Consul KV). Returns live status and API base URLs for all registered modules (sql, compute, iam, etc.).",
		Attributes: map[string]schema.Attribute{
			"services": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of registered platform services.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Service name (e.g. sql, compute, iam).",
						},
						"endpoint": schema.StringAttribute{
							Computed:    true,
							Description: "API base URL for this service.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Health status: healthy, degraded, or unhealthy.",
						},
						"version": schema.StringAttribute{
							Computed:    true,
							Description: "Deployed version of the service.",
						},
					},
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the service registry was last updated (ISO UTC).",
			},
		},
	}
}

func (d *serviceDiscoveryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceDiscoveryDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	sd, err := d.client.GetIAMServiceDiscovery(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service discovery", err.Error())
		return
	}

	objs := make([]attr.Value, len(sd.Services))
	for i, svc := range sd.Services {
		obj, diags := types.ObjectValue(serviceAttrTypes, map[string]attr.Value{
			"name":     types.StringValue(svc.Name),
			"endpoint": types.StringValue(svc.Endpoint),
			"status":   types.StringValue(svc.Status),
			"version":  types.StringValue(svc.Version),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: serviceAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedAt := types.StringNull()
	if sd.UpdatedAt != nil {
		updatedAt = types.StringValue(*sd.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, serviceDiscoveryModel{
		Services:  list,
		UpdatedAt: updatedAt,
	})...)
}
