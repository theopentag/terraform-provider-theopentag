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

var _ datasource.DataSource = &backupsDataSource{}

type backupsDataSource struct {
	client *client.Client
}

type backupsModel struct {
	ServerName types.String `tfsdk:"server_name"`
	Backups    types.List   `tfsdk:"backups"`
}

var backupAttrTypes = map[string]attr.Type{
	"backup_id":   types.StringType,
	"status":      types.StringType,
	"size":        types.StringType,
	"begin_time":  types.StringType,
	"end_time":    types.StringType,
	"backup_type": types.StringType,
	"source":      types.StringType,
}

func NewBackupsDataSource() datasource.DataSource {
	return &backupsDataSource{}
}

func (d *backupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_backups"
}

func (d *backupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all backups for a SQL-managed PostgreSQL server.",
		Attributes: map[string]schema.Attribute{
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SQL-managed server.",
			},
			"backups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of backups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"backup_id": schema.StringAttribute{
							Computed:    true,
							Description: "Backup identifier (e.g. 20240101T120000).",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Backup status: DONE, FAILED, STARTED, WAITING_FOR_WALS, DONE_WITH_ERRORS, or EMPTY.",
						},
						"size": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable backup size.",
						},
						"begin_time": schema.StringAttribute{
							Computed:    true,
							Description: "Backup start time.",
						},
						"end_time": schema.StringAttribute{
							Computed:    true,
							Description: "Backup end time.",
						},
						"backup_type": schema.StringAttribute{
							Computed:    true,
							Description: "Backup type.",
						},
						"source": schema.StringAttribute{
							Computed:    true,
							Description: "Backup source: manual or scheduler.",
						},
					},
				},
			},
		},
	}
}

func (d *backupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *backupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config backupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backups, err := d.client.ListBackups(ctx, config.ServerName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing backups", err.Error())
		return
	}

	backupObjs := make([]attr.Value, len(backups))
	for i, b := range backups {
		beginTime := types.StringNull()
		if b.BeginTime != nil {
			beginTime = types.StringValue(*b.BeginTime)
		}
		endTime := types.StringNull()
		if b.EndTime != nil {
			endTime = types.StringValue(*b.EndTime)
		}
		backupType := types.StringNull()
		if b.BackupType != nil {
			backupType = types.StringValue(*b.BackupType)
		}
		source := types.StringNull()
		if b.Source != nil {
			source = types.StringValue(*b.Source)
		}

		obj, diags := types.ObjectValue(backupAttrTypes, map[string]attr.Value{
			"backup_id":   types.StringValue(b.BackupID),
			"status":      types.StringValue(b.Status),
			"size":        types.StringValue(b.Size),
			"begin_time":  beginTime,
			"end_time":    endTime,
			"backup_type": backupType,
			"source":      source,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		backupObjs[i] = obj
	}

	backupsList, diags := types.ListValue(types.ObjectType{AttrTypes: backupAttrTypes}, backupObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := backupsModel{
		ServerName: config.ServerName,
		Backups:    backupsList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
