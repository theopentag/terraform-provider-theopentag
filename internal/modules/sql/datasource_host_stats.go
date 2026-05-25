package sql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &hostStatsDataSource{}

type hostStatsDataSource struct {
	client *client.Client
}

type hostStatsModel struct {
	CPUPercent  types.Float64 `tfsdk:"cpu_percent"`
	RAMTotal    types.Int64   `tfsdk:"ram_total"`
	RAMUsed     types.Int64   `tfsdk:"ram_used"`
	RAMPercent  types.Float64 `tfsdk:"ram_percent"`
	DiskTotal   types.Int64   `tfsdk:"disk_total"`
	DiskUsed    types.Int64   `tfsdk:"disk_used"`
	DiskPercent types.Float64 `tfsdk:"disk_percent"`
	Timestamp   types.String  `tfsdk:"timestamp"`
}

func NewHostStatsDataSource() datasource.DataSource {
	return &hostStatsDataSource{}
}

func (d *hostStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_host_stats"
}

func (d *hostStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Latest backend host CPU, RAM, and disk snapshot. Errors if no data has been collected yet.",
		Attributes: map[string]schema.Attribute{
			"cpu_percent": schema.Float64Attribute{
				Computed:    true,
				Description: "CPU usage percent (rate over the last 15s window).",
			},
			"ram_total": schema.Int64Attribute{
				Computed:    true,
				Description: "Total RAM in bytes.",
			},
			"ram_used": schema.Int64Attribute{
				Computed:    true,
				Description: "Used RAM in bytes.",
			},
			"ram_percent": schema.Float64Attribute{
				Computed:    true,
				Description: "RAM usage percent.",
			},
			"disk_total": schema.Int64Attribute{
				Computed:    true,
				Description: "Total disk size in bytes (backend container root filesystem).",
			},
			"disk_used": schema.Int64Attribute{
				Computed:    true,
				Description: "Used disk in bytes.",
			},
			"disk_percent": schema.Float64Attribute{
				Computed:    true,
				Description: "Disk usage percent.",
			},
			"timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of this snapshot (ISO UTC).",
			},
		},
	}
}

func (d *hostStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hostStatsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	h, err := d.client.GetHostStats(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading host stats", err.Error())
		return
	}
	if h == nil {
		resp.Diagnostics.AddError("No host stats available", "The backend has not collected any host stats yet. Retry after a few seconds.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, hostStatsModel{
		CPUPercent:  types.Float64Value(h.CPUPercent),
		RAMTotal:    types.Int64Value(h.RAMTotal),
		RAMUsed:     types.Int64Value(h.RAMUsed),
		RAMPercent:  types.Float64Value(h.RAMPercent),
		DiskTotal:   types.Int64Value(h.DiskTotal),
		DiskUsed:    types.Int64Value(h.DiskUsed),
		DiskPercent: types.Float64Value(h.DiskPercent),
		Timestamp:   types.StringValue(h.Timestamp),
	})...)
}
