package sql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ resource.Resource = &serverConfigResource{}
var _ resource.ResourceWithImportState = &serverConfigResource{}

type serverConfigResource struct {
	client *client.Client
}

type serverConfigModel struct {
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	Conninfo                   types.String `tfsdk:"conninfo"`
	SSHCommand                 types.String `tfsdk:"ssh_command"`
	BackupMethod               types.String `tfsdk:"backup_method"`
	Archiver                   types.Bool   `tfsdk:"archiver"`
	StreamingConninfo          types.String `tfsdk:"streaming_conninfo"`
	StreamingArchiver          types.Bool   `tfsdk:"streaming_archiver"`
	CreateSlot                 types.String `tfsdk:"create_slot"`
	SlotName                   types.String `tfsdk:"slot_name"`
	PathPrefix                 types.String `tfsdk:"path_prefix"`
	SSLMode                    types.String `tfsdk:"sslmode"`
	RetentionPolicy            types.String `tfsdk:"retention_policy"`
	WALRetentionPolicy         types.String `tfsdk:"wal_retention_policy"`
	MinimumRedundancy          types.Int64  `tfsdk:"minimum_redundancy"`
	Compression                types.String `tfsdk:"compression"`
	BackupCompression          types.String `tfsdk:"backup_compression"`
	StreamingArchiverBatchSize types.Int64  `tfsdk:"streaming_archiver_batch_size"`
	PGVersion                  types.Int64  `tfsdk:"pg_version"`
	BackupsEnabled             types.Bool   `tfsdk:"backups_enabled"`
	ScheduleEnabled            types.Bool   `tfsdk:"schedule_enabled"`
}

func NewServerConfigResource() resource.Resource {
	return &serverConfigResource{}
}

func (r *serverConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_server_config"
}

func (r *serverConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers a PostgreSQL server with the SQL management system and configures its backup settings.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique server name within the SQL system ([a-zA-Z0-9_-]+). Immutable — changes force replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"conninfo": schema.StringAttribute{
				Required:    true,
				Description: "libpq connection string (may include password=).",
			},
			"ssh_command": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH command for rsync backup method.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backup_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Backup method: postgres or rsync.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"archiver": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable WAL archiver.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"streaming_conninfo": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Streaming replication connection string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"streaming_archiver": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable streaming archiver.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"create_slot": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Replication slot creation mode: auto or manual.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slot_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Replication slot name. Defaults to server name with hyphens replaced by underscores.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path_prefix": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path to PostgreSQL binaries (e.g. /usr/lib/postgresql/17/bin/).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sslmode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSL mode appended to conninfo and streaming_conninfo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Backup retention policy (e.g. RECOVERY WINDOW OF 14 DAYS).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wal_retention_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "WAL retention policy. Always 'main'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"minimum_redundancy": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Minimum number of backups to retain.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"compression": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "WAL compression algorithm (e.g. bzip2).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backup_compression": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Backup compression algorithm (e.g. gzip).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"streaming_archiver_batch_size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of WAL files to retrieve per streaming archiver batch.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"pg_version": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "PostgreSQL major version (14–18).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"backups_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether backups are enabled for this server.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to enable the auto-created daily backup schedule on server creation. Only consumed on create; not tracked after.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serverConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedEnabled := true
	if !plan.ScheduleEnabled.IsNull() && !plan.ScheduleEnabled.IsUnknown() {
		schedEnabled = plan.ScheduleEnabled.ValueBool()
	}

	createReq := client.ServerConfigCreateRequest{
		ServerConfig:    modelToServerConfig(plan),
		ScheduleEnabled: &schedEnabled,
	}
	createReq.ServerConfig.Name = plan.Name.ValueString()

	sc, err := r.client.CreateServerConfig(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server config", err.Error())
		return
	}

	state := serverConfigToModel(sc)
	state.ScheduleEnabled = types.BoolValue(schedEnabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sc, err := r.client.GetServerConfig(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading server config", err.Error())
		return
	}
	if sc == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := serverConfigToModel(sc)
	newState.ScheduleEnabled = state.ScheduleEnabled
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *serverConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serverConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modelToServerConfig(plan)
	body.Name = plan.Name.ValueString()
	sc, err := r.client.UpdateServerConfig(ctx, plan.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server config", err.Error())
		return
	}

	if !plan.BackupsEnabled.IsNull() && !plan.BackupsEnabled.IsUnknown() &&
		plan.BackupsEnabled.ValueBool() != state.BackupsEnabled.ValueBool() {
		if err := r.client.SetBackupsEnabled(ctx, plan.Name.ValueString(), plan.BackupsEnabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error setting backups_enabled", err.Error())
			return
		}
		sc.BackupsEnabled = plan.BackupsEnabled.ValueBool()
	}

	newState := serverConfigToModel(sc)
	newState.ScheduleEnabled = state.ScheduleEnabled
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *serverConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServerConfig(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting server config", err.Error())
	}
}

func modelToServerConfig(m serverConfigModel) client.ServerConfig {
	sc := client.ServerConfig{
		Conninfo:           m.Conninfo.ValueString(),
		BackupMethod:       m.BackupMethod.ValueString(),
		Archiver:           m.Archiver.ValueBool(),
		StreamingArchiver:  m.StreamingArchiver.ValueBool(),
		CreateSlot:         m.CreateSlot.ValueString(),
		SSLMode:            m.SSLMode.ValueString(),
		WALRetentionPolicy: m.WALRetentionPolicy.ValueString(),
		MinimumRedundancy:  m.MinimumRedundancy.ValueInt64(),
		StreamingArchiverBatchSize: m.StreamingArchiverBatchSize.ValueInt64(),
		PGVersion:          m.PGVersion.ValueInt64(),
		BackupsEnabled:     m.BackupsEnabled.ValueBool(),
	}

	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		sc.Description = &v
	}
	if !m.SSHCommand.IsNull() && !m.SSHCommand.IsUnknown() {
		v := m.SSHCommand.ValueString()
		sc.SSHCommand = &v
	}
	if !m.StreamingConninfo.IsNull() && !m.StreamingConninfo.IsUnknown() {
		v := m.StreamingConninfo.ValueString()
		sc.StreamingConninfo = &v
	}
	if !m.SlotName.IsNull() && !m.SlotName.IsUnknown() {
		v := m.SlotName.ValueString()
		sc.SlotName = &v
	}
	if !m.PathPrefix.IsNull() && !m.PathPrefix.IsUnknown() {
		v := m.PathPrefix.ValueString()
		sc.PathPrefix = &v
	}
	if !m.RetentionPolicy.IsNull() && !m.RetentionPolicy.IsUnknown() {
		v := m.RetentionPolicy.ValueString()
		sc.RetentionPolicy = &v
	}
	if !m.Compression.IsNull() && !m.Compression.IsUnknown() {
		v := m.Compression.ValueString()
		sc.Compression = &v
	}
	if !m.BackupCompression.IsNull() && !m.BackupCompression.IsUnknown() {
		v := m.BackupCompression.ValueString()
		sc.BackupCompression = &v
	}

	return sc
}

func serverConfigToModel(sc *client.ServerConfig) serverConfigModel {
	m := serverConfigModel{
		Name:                       types.StringValue(sc.Name),
		Conninfo:                   types.StringValue(sc.Conninfo),
		BackupMethod:               types.StringValue(sc.BackupMethod),
		Archiver:                   types.BoolValue(sc.Archiver),
		StreamingArchiver:          types.BoolValue(sc.StreamingArchiver),
		CreateSlot:                 types.StringValue(sc.CreateSlot),
		SSLMode:                    types.StringValue(sc.SSLMode),
		WALRetentionPolicy:         types.StringValue(sc.WALRetentionPolicy),
		MinimumRedundancy:          types.Int64Value(sc.MinimumRedundancy),
		StreamingArchiverBatchSize: types.Int64Value(sc.StreamingArchiverBatchSize),
		PGVersion:                  types.Int64Value(sc.PGVersion),
		BackupsEnabled:             types.BoolValue(sc.BackupsEnabled),
	}

	if sc.Description != nil {
		m.Description = types.StringValue(*sc.Description)
	} else {
		m.Description = types.StringNull()
	}
	if sc.SSHCommand != nil {
		m.SSHCommand = types.StringValue(*sc.SSHCommand)
	} else {
		m.SSHCommand = types.StringNull()
	}
	if sc.StreamingConninfo != nil {
		m.StreamingConninfo = types.StringValue(*sc.StreamingConninfo)
	} else {
		m.StreamingConninfo = types.StringNull()
	}
	if sc.SlotName != nil {
		m.SlotName = types.StringValue(*sc.SlotName)
	} else {
		m.SlotName = types.StringNull()
	}
	if sc.PathPrefix != nil {
		m.PathPrefix = types.StringValue(*sc.PathPrefix)
	} else {
		m.PathPrefix = types.StringNull()
	}
	if sc.RetentionPolicy != nil {
		m.RetentionPolicy = types.StringValue(*sc.RetentionPolicy)
	} else {
		m.RetentionPolicy = types.StringNull()
	}
	if sc.Compression != nil {
		m.Compression = types.StringValue(*sc.Compression)
	} else {
		m.Compression = types.StringNull()
	}
	if sc.BackupCompression != nil {
		m.BackupCompression = types.StringValue(*sc.BackupCompression)
	} else {
		m.BackupCompression = types.StringNull()
	}

	return m
}

func (r *serverConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
