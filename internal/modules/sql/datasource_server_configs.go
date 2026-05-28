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

var _ datasource.DataSource = &serverConfigsDataSource{}

type serverConfigsDataSource struct {
	client *client.Client
}

type serverConfigsModel struct {
	ServerConfigs types.List `tfsdk:"server_configs"`
}

var serverConfigAttrTypes = map[string]attr.Type{
	"name":                         types.StringType,
	"description":                  types.StringType,
	"conninfo":                     types.StringType,
	"ssh_command":                  types.StringType,
	"backup_method":                types.StringType,
	"archiver":                     types.BoolType,
	"streaming_conninfo":           types.StringType,
	"streaming_archiver":           types.BoolType,
	"create_slot":                  types.StringType,
	"slot_name":                    types.StringType,
	"path_prefix":                  types.StringType,
	"sslmode":                      types.StringType,
	"retention_policy":             types.StringType,
	"wal_retention_policy":         types.StringType,
	"minimum_redundancy":           types.Int64Type,
	"compression":                  types.StringType,
	"backup_compression":           types.StringType,
	"streaming_archiver_batch_size": types.Int64Type,
	"pg_version":                   types.Int64Type,
	"backups_enabled":              types.BoolType,
}

func NewServerConfigsDataSource() datasource.DataSource {
	return &serverConfigsDataSource{}
}

func (d *serverConfigsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_server_configs"
}

func (d *serverConfigsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all servers registered in the SQL management system.",
		Attributes: map[string]schema.Attribute{
			"server_configs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of server configurations.",
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
						"conninfo": schema.StringAttribute{
							Computed:    true,
							Description: "libpq connection string.",
						},
						"ssh_command": schema.StringAttribute{
							Computed:    true,
							Description: "SSH command for rsync backup method.",
						},
						"backup_method": schema.StringAttribute{
							Computed:    true,
							Description: "Backup method: postgres or rsync.",
						},
						"archiver": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether WAL archiver is enabled.",
						},
						"streaming_conninfo": schema.StringAttribute{
							Computed:    true,
							Description: "Streaming replication connection string.",
						},
						"streaming_archiver": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether streaming archiver is enabled.",
						},
						"create_slot": schema.StringAttribute{
							Computed:    true,
							Description: "Replication slot creation mode: auto or manual.",
						},
						"slot_name": schema.StringAttribute{
							Computed:    true,
							Description: "Replication slot name.",
						},
						"path_prefix": schema.StringAttribute{
							Computed:    true,
							Description: "Path to PostgreSQL binaries.",
						},
						"sslmode": schema.StringAttribute{
							Computed:    true,
							Description: "SSL mode appended to conninfo and streaming_conninfo.",
						},
						"retention_policy": schema.StringAttribute{
							Computed:    true,
							Description: "Backup retention policy.",
						},
						"wal_retention_policy": schema.StringAttribute{
							Computed:    true,
							Description: "WAL retention policy. Always 'main'.",
						},
						"minimum_redundancy": schema.Int64Attribute{
							Computed:    true,
							Description: "Minimum number of backups to retain.",
						},
						"compression": schema.StringAttribute{
							Computed:    true,
							Description: "WAL compression algorithm.",
						},
						"backup_compression": schema.StringAttribute{
							Computed:    true,
							Description: "Backup compression algorithm.",
						},
						"streaming_archiver_batch_size": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of WAL files to retrieve per streaming archiver batch.",
						},
						"pg_version": schema.Int64Attribute{
							Computed:    true,
							Description: "PostgreSQL major version (14–18).",
						},
						"backups_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether backups are enabled for this server.",
						},
					},
				},
			},
		},
	}
}

func (d *serverConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	configs, err := d.client.ListServerConfigs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing server configs", err.Error())
		return
	}

	objs := make([]attr.Value, len(configs))
	for i, sc := range configs {
		m := serverConfigToModel(&sc)

		obj, diags := types.ObjectValue(serverConfigAttrTypes, map[string]attr.Value{
			"name":                         types.StringValue(sc.Name),
			"description":                  m.Description,
			"conninfo":                     types.StringValue(sc.Conninfo),
			"ssh_command":                  m.SSHCommand,
			"backup_method":                types.StringValue(sc.BackupMethod),
			"archiver":                     types.BoolValue(bool(sc.Archiver)),
			"streaming_conninfo":           m.StreamingConninfo,
			"streaming_archiver":           types.BoolValue(bool(sc.StreamingArchiver)),
			"create_slot":                  types.StringValue(sc.CreateSlot),
			"slot_name":                    m.SlotName,
			"path_prefix":                  m.PathPrefix,
			"sslmode":                      types.StringValue(sc.SSLMode),
			"retention_policy":             m.RetentionPolicy,
			"wal_retention_policy":         types.StringValue(sc.WALRetentionPolicy),
			"minimum_redundancy":           types.Int64Value(sc.MinimumRedundancy),
			"compression":                  m.Compression,
			"backup_compression":           m.BackupCompression,
			"streaming_archiver_batch_size": types.Int64Value(sc.StreamingArchiverBatchSize),
			"pg_version":                   types.Int64Value(sc.PGVersion),
			"backups_enabled":              types.BoolValue(bool(sc.BackupsEnabled)),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: serverConfigAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, serverConfigsModel{ServerConfigs: list})...)
}
