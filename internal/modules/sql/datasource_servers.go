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

var _ datasource.DataSource = &serversDataSource{}

type serversDataSource struct {
	client *client.Client
}

type serversModel struct {
	Servers types.List `tfsdk:"servers"`
}

var serverSummaryAttrTypes = map[string]attr.Type{
	"name":              types.StringType,
	"description":       types.StringType,
	"active":            types.BoolType,
	"backup_count":      types.Int64Type,
	"last_backup":       types.StringType,
	"disk_usage":        types.StringType,
	"disk_bytes":        types.Int64Type,
	"retention_policy":  types.StringType,
	"check_ok":          types.BoolType,
	"redundancy_ok":     types.BoolType,
	"redundancy_raw":    types.StringType,
	"has_active_backup": types.BoolType,
}

func NewServersDataSource() datasource.DataSource {
	return &serversDataSource{}
}

func (d *serversDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_servers"
}

func (d *serversDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all PostgreSQL servers registered in the SQL management system with their current status.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of servers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Server name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable description.",
						},
						"active": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether barman considers the server active.",
						},
						"backup_count": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of available backups.",
						},
						"last_backup": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of the most recent backup (ISO UTC).",
						},
						"disk_usage": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable backup disk usage.",
						},
						"disk_bytes": schema.Int64Attribute{
							Computed:    true,
							Description: "Backup disk usage in bytes.",
						},
						"retention_policy": schema.StringAttribute{
							Computed:    true,
							Description: "Configured retention policy string.",
						},
						"check_ok": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether all critical checks pass.",
						},
						"redundancy_ok": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether minimum redundancy is met.",
						},
						"redundancy_raw": schema.StringAttribute{
							Computed:    true,
							Description: "Raw redundancy string from barman.",
						},
						"has_active_backup": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether a backup is currently in progress.",
						},
					},
				},
			},
		},
	}
}

func (d *serversDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	servers, err := d.client.ListServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing servers", err.Error())
		return
	}

	objs := make([]attr.Value, len(servers))
	for i, s := range servers {
		description := types.StringNull()
		if s.Description != nil {
			description = types.StringValue(*s.Description)
		}
		lastBackup := types.StringNull()
		if s.LastBackup != nil {
			lastBackup = types.StringValue(*s.LastBackup)
		}
		diskUsage := types.StringNull()
		if s.DiskUsage != nil {
			diskUsage = types.StringValue(*s.DiskUsage)
		}
		diskBytes := types.Int64Null()
		if s.DiskBytes != nil {
			diskBytes = types.Int64Value(*s.DiskBytes)
		}
		retentionPolicy := types.StringNull()
		if s.RetentionPolicy != nil {
			retentionPolicy = types.StringValue(*s.RetentionPolicy)
		}
		redundancyRaw := types.StringNull()
		if s.RedundancyRaw != nil {
			redundancyRaw = types.StringValue(*s.RedundancyRaw)
		}

		obj, diags := types.ObjectValue(serverSummaryAttrTypes, map[string]attr.Value{
			"name":              types.StringValue(s.Name),
			"description":       description,
			"active":            types.BoolValue(s.Active),
			"backup_count":      types.Int64Value(s.BackupCount),
			"last_backup":       lastBackup,
			"disk_usage":        diskUsage,
			"disk_bytes":        diskBytes,
			"retention_policy":  retentionPolicy,
			"check_ok":          types.BoolValue(s.CheckOK),
			"redundancy_ok":     types.BoolValue(s.RedundancyOK),
			"redundancy_raw":    redundancyRaw,
			"has_active_backup": types.BoolValue(s.HasActiveBackup),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: serverSummaryAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, serversModel{Servers: list})...)
}
