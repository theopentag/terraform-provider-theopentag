package sql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

var _ resource.Resource = &scheduleResource{}
var _ resource.ResourceWithImportState = &scheduleResource{}

type scheduleResource struct {
	client *client.Client
}

type scheduleModel struct {
	ID             types.String        `tfsdk:"id"`
	ServerName     types.String        `tfsdk:"server_name"`
	Label          types.String        `tfsdk:"label"`
	ScheduleType   types.String        `tfsdk:"schedule_type"`
	ScheduleConfig scheduleConfigModel `tfsdk:"schedule_config"`
	Enabled        types.Bool          `tfsdk:"enabled"`
	NextRunAt      types.String        `tfsdk:"next_run_at"`
	LastRunAt      types.String        `tfsdk:"last_run_at"`
	CreatedAt      types.String        `tfsdk:"created_at"`
}

type scheduleConfigModel struct {
	RunAt types.String `tfsdk:"run_at"`
	Time  types.String `tfsdk:"time"`
	Days  types.List   `tfsdk:"days"`
	Day   types.Int64  `tfsdk:"day"`
}

func NewScheduleResource() resource.Resource {
	return &scheduleResource{}
}

func (r *scheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_schedule"
}

func (r *scheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backup schedule for a SQL-managed PostgreSQL server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Schedule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SQL-managed server this schedule applies to.",
			},
			"label": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable label for the schedule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_type": schema.StringAttribute{
				Required:    true,
				Description: "Schedule type: once, daily, weekly, or monthly.",
			},
			"schedule_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Schedule configuration. Content depends on schedule_type.",
				Attributes: map[string]schema.Attribute{
					"run_at": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "ISO datetime for once-type schedules.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"time": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Time in HH:MM (UTC) for daily/weekly/monthly schedules.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"days": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Days of the week (0=Sun–6=Sat) for weekly schedules.",
					},
					"day": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Day of the month (1–31) for monthly schedules.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the schedule is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"next_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "Next scheduled run time (ISO UTC). Recomputed by the server after every update.",
			},
			"last_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last run time (ISO UTC).",
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

func (r *scheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sreq, d := planToScheduleRequest(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.CreateSchedule(ctx, sreq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating schedule", err.Error())
		return
	}

	state, d := scheduleToModel(ctx, s)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *scheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid schedule ID", err.Error())
		return
	}

	s, err := r.client.GetSchedule(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading schedule", err.Error())
		return
	}
	if s == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, d := scheduleToModel(ctx, s)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *scheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid schedule ID", err.Error())
		return
	}

	sreq, d := planToScheduleRequest(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateSchedule(ctx, id, sreq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating schedule", err.Error())
		return
	}

	state, d := scheduleToModel(ctx, s)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *scheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid schedule ID", err.Error())
		return
	}

	if err := r.client.DeleteSchedule(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error deleting schedule", err.Error())
	}
}

func planToScheduleRequest(ctx context.Context, m scheduleModel) (client.ScheduleRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	sreq := client.ScheduleRequest{
		ServerName:   m.ServerName.ValueString(),
		ScheduleType: m.ScheduleType.ValueString(),
		Enabled:      m.Enabled.ValueBool(),
	}

	if !m.Label.IsNull() && !m.Label.IsUnknown() {
		v := m.Label.ValueString()
		sreq.Label = &v
	}

	cfg := client.ScheduleConfig{}
	sc := m.ScheduleConfig

	if !sc.RunAt.IsNull() && !sc.RunAt.IsUnknown() {
		v := sc.RunAt.ValueString()
		cfg.RunAt = &v
	}
	if !sc.Time.IsNull() && !sc.Time.IsUnknown() {
		v := sc.Time.ValueString()
		cfg.Time = &v
	}
	if !sc.Days.IsNull() && !sc.Days.IsUnknown() {
		var days []int64
		d := sc.Days.ElementsAs(ctx, &days, false)
		diags.Append(d...)
		if !d.HasError() {
			cfg.Days = days
		}
	}
	if !sc.Day.IsNull() && !sc.Day.IsUnknown() {
		v := sc.Day.ValueInt64()
		cfg.Day = &v
	}

	sreq.ScheduleConfig = cfg
	return sreq, diags
}

func scheduleToModel(ctx context.Context, s *client.Schedule) (scheduleModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	m := scheduleModel{
		ID:           types.StringValue(strconv.FormatInt(s.ID, 10)),
		ServerName:   types.StringValue(s.ServerName),
		ScheduleType: types.StringValue(s.ScheduleType),
		Enabled:      types.BoolValue(s.Enabled),
		CreatedAt:    types.StringValue(s.CreatedAt),
	}

	if s.Label != nil {
		m.Label = types.StringValue(*s.Label)
	} else {
		m.Label = types.StringNull()
	}
	if s.NextRunAt != nil {
		m.NextRunAt = types.StringValue(*s.NextRunAt)
	} else {
		m.NextRunAt = types.StringNull()
	}
	if s.LastRunAt != nil {
		m.LastRunAt = types.StringValue(*s.LastRunAt)
	} else {
		m.LastRunAt = types.StringNull()
	}

	sc := s.ScheduleConfig
	cfgModel := scheduleConfigModel{}

	if sc.RunAt != nil {
		cfgModel.RunAt = types.StringValue(*sc.RunAt)
	} else {
		cfgModel.RunAt = types.StringNull()
	}
	if sc.Time != nil {
		cfgModel.Time = types.StringValue(*sc.Time)
	} else {
		cfgModel.Time = types.StringNull()
	}

	if len(sc.Days) > 0 {
		elems := make([]attr.Value, len(sc.Days))
		for i, d := range sc.Days {
			elems[i] = types.Int64Value(d)
		}
		list, d := types.ListValue(types.Int64Type, elems)
		diags.Append(d...)
		cfgModel.Days = list
	} else {
		cfgModel.Days = types.ListValueMust(types.Int64Type, []attr.Value{})
	}

	if sc.Day != nil {
		cfgModel.Day = types.Int64Value(*sc.Day)
	} else {
		cfgModel.Day = types.Int64Null()
	}

	m.ScheduleConfig = cfgModel
	return m, diags
}

func (r *scheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

