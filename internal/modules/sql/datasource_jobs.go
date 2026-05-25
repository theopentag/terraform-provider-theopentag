package sql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &jobsDataSource{}

type jobsDataSource struct {
	client *client.Client
}

type jobsModel struct {
	Server types.String `tfsdk:"server"`
	Status types.String `tfsdk:"status"`
	Limit  types.Int64  `tfsdk:"limit"`
	Jobs   types.List   `tfsdk:"jobs"`
}

var jobAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"cache_key":    types.StringType,
	"args_json":    types.StringType,
	"status":       types.StringType,
	"exit_code":    types.StringType,
	"stdout":       types.StringType,
	"stderr":       types.StringType,
	"queued_at":    types.StringType,
	"started_at":   types.StringType,
	"completed_at": types.StringType,
	"schedule_id":  types.StringType,
	"pg_version":   types.StringType,
}

func NewJobsDataSource() datasource.DataSource {
	return &jobsDataSource{}
}

func (d *jobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_jobs"
}

func (d *jobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists SQL command queue jobs.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Optional:    true,
				Description: "Filter jobs by server name.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter jobs by status: pending, running, done, or failed.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of jobs to return (1–200, default 50).",
			},
			"jobs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of jobs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Job ID.",
						},
						"cache_key": schema.StringAttribute{
							Computed:    true,
							Description: "Deduplication cache key.",
						},
						"args_json": schema.StringAttribute{
							Computed:    true,
							Description: "JSON array of command arguments.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Job status: pending, running, done, or failed.",
						},
						"exit_code": schema.StringAttribute{
							Computed:    true,
							Description: "Process exit code (as string, null if not yet completed).",
						},
						"stdout": schema.StringAttribute{
							Computed:    true,
							Description: "Standard output.",
						},
						"stderr": schema.StringAttribute{
							Computed:    true,
							Description: "Standard error.",
						},
						"queued_at": schema.StringAttribute{
							Computed:    true,
							Description: "Time the job was queued (ISO UTC).",
						},
						"started_at": schema.StringAttribute{
							Computed:    true,
							Description: "Time the job started (ISO UTC).",
						},
						"completed_at": schema.StringAttribute{
							Computed:    true,
							Description: "Time the job completed (ISO UTC).",
						},
						"schedule_id": schema.StringAttribute{
							Computed:    true,
							Description: "Schedule ID that triggered this job, if any.",
						},
						"pg_version": schema.StringAttribute{
							Computed:    true,
							Description: "PostgreSQL major version this job targets, if any.",
						},
					},
				},
			},
		},
	}
}

func (d *jobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *jobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config jobsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := ""
	if !config.Server.IsNull() && !config.Server.IsUnknown() {
		server = config.Server.ValueString()
	}
	statusFilter := ""
	if !config.Status.IsNull() && !config.Status.IsUnknown() {
		statusFilter = config.Status.ValueString()
	}
	limit := int64(50)
	if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		limit = config.Limit.ValueInt64()
	}

	jobs, err := d.client.ListJobs(ctx, server, statusFilter, limit)
	if err != nil {
		resp.Diagnostics.AddError("Error listing jobs", err.Error())
		return
	}

	jobObjs := make([]attr.Value, len(jobs))
	for i, j := range jobs {
		cacheKey := types.StringNull()
		if j.CacheKey != nil {
			cacheKey = types.StringValue(*j.CacheKey)
		}
		exitCode := types.StringNull()
		if j.ExitCode != nil {
			exitCode = types.StringValue(strconv.FormatInt(*j.ExitCode, 10))
		}
		stdout := types.StringNull()
		if j.Stdout != nil {
			stdout = types.StringValue(*j.Stdout)
		}
		stderr := types.StringNull()
		if j.Stderr != nil {
			stderr = types.StringValue(*j.Stderr)
		}
		startedAt := types.StringNull()
		if j.StartedAt != nil {
			startedAt = types.StringValue(*j.StartedAt)
		}
		completedAt := types.StringNull()
		if j.CompletedAt != nil {
			completedAt = types.StringValue(*j.CompletedAt)
		}
		scheduleID := types.StringNull()
		if j.ScheduleID != nil {
			scheduleID = types.StringValue(strconv.FormatInt(*j.ScheduleID, 10))
		}
		pgVersion := types.StringNull()
		if j.PGVersion != nil {
			pgVersion = types.StringValue(strconv.FormatInt(*j.PGVersion, 10))
		}

		obj, diags := types.ObjectValue(jobAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(strconv.FormatInt(j.ID, 10)),
			"cache_key":    cacheKey,
			"args_json":    types.StringValue(j.ArgsJSON),
			"status":       types.StringValue(j.Status),
			"exit_code":    exitCode,
			"stdout":       stdout,
			"stderr":       stderr,
			"queued_at":    types.StringValue(j.QueuedAt),
			"started_at":   startedAt,
			"completed_at": completedAt,
			"schedule_id":  scheduleID,
			"pg_version":   pgVersion,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		jobObjs[i] = obj
	}

	jobsList, diags := types.ListValue(types.ObjectType{AttrTypes: jobAttrTypes}, jobObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := jobsModel{
		Server: config.Server,
		Status: config.Status,
		Limit:  config.Limit,
		Jobs:   jobsList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
