package sql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &statsDataSource{}

type statsDataSource struct {
	client *client.Client
}

type statsModel struct {
	TotalBackups    types.Int64  `tfsdk:"total_backups"`
	TotalDisk       types.String `tfsdk:"total_disk"`
	BarmanDiskTotal types.Int64  `tfsdk:"barman_disk_total"`
	BarmanDiskFree  types.Int64  `tfsdk:"barman_disk_free"`
}

func NewStatsDataSource() datasource.DataSource {
	return &statsDataSource{}
}

func (d *statsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_stats"
}

func (d *statsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Aggregate backup statistics across all servers.",
		Attributes: map[string]schema.Attribute{
			"total_backups": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of backups across all servers.",
			},
			"total_disk": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable total backup disk usage.",
			},
			"barman_disk_total": schema.Int64Attribute{
				Computed:    true,
				Description: "Total barman data directory filesystem size in bytes.",
			},
			"barman_disk_free": schema.Int64Attribute{
				Computed:    true,
				Description: "Free space on the barman data directory filesystem in bytes.",
			},
		},
	}
}

func (d *statsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	s, err := d.client.GetStats(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading stats", err.Error())
		return
	}

	state := statsModel{
		TotalBackups: types.Int64Value(s.TotalBackups),
		TotalDisk:    types.StringValue(s.TotalDisk),
	}
	if s.BarmanDiskTotal != nil {
		state.BarmanDiskTotal = types.Int64Value(*s.BarmanDiskTotal)
	} else {
		state.BarmanDiskTotal = types.Int64Null()
	}
	if s.BarmanDiskFree != nil {
		state.BarmanDiskFree = types.Int64Value(*s.BarmanDiskFree)
	} else {
		state.BarmanDiskFree = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
