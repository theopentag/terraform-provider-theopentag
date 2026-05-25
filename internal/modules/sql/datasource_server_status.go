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

var _ datasource.DataSource = &serverStatusDataSource{}

type serverStatusDataSource struct {
	client *client.Client
}

type serverStatusModel struct {
	ServerName      types.String `tfsdk:"server_name"`
	OK              types.Bool   `tfsdk:"ok"`
	CheckOK         types.Bool   `tfsdk:"check_ok"`
	CheckItems      types.List   `tfsdk:"check_items"`
	Fields          types.Map    `tfsdk:"fields"`
	ReplicationJSON types.String `tfsdk:"replication_json"`
}

type checkItemModel struct {
	Check  types.String `tfsdk:"check"`
	Status types.String `tfsdk:"status"`
	Hint   types.String `tfsdk:"hint"`
}

var checkItemAttrTypes = map[string]attr.Type{
	"check":  types.StringType,
	"status": types.StringType,
	"hint":   types.StringType,
}

func NewServerStatusDataSource() datasource.DataSource {
	return &serverStatusDataSource{}
}

func (d *serverStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_server_status"
}

func (d *serverStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the current status of a SQL-managed PostgreSQL server.",
		Attributes: map[string]schema.Attribute{
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SQL-managed server.",
			},
			"ok": schema.BoolAttribute{
				Computed:    true,
				Description: "Overall server health.",
			},
			"check_ok": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether all critical checks passed.",
			},
			"check_items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of individual check results.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"check": schema.StringAttribute{
							Computed:    true,
							Description: "Check name.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Check status: OK or FAILED.",
						},
						"hint": schema.StringAttribute{
							Computed:    true,
							Description: "Additional hint or description.",
						},
					},
				},
			},
			"fields": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "All other barman status fields as a string map.",
			},
			"replication_json": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Replication status as a raw JSON string.",
			},
		},
	}
}

func (d *serverStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serverStatusModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ss, err := d.client.GetServerStatus(ctx, config.ServerName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading server status", err.Error())
		return
	}

	state := serverStatusModel{
		ServerName: config.ServerName,
		OK:         types.BoolValue(ss.OK),
		CheckOK:    types.BoolValue(ss.CheckOK),
	}

	checkItemObjs := make([]attr.Value, len(ss.CheckItems))
	for i, ci := range ss.CheckItems {
		obj, diags := types.ObjectValue(checkItemAttrTypes, map[string]attr.Value{
			"check":  types.StringValue(ci.Check),
			"status": types.StringValue(ci.Status),
			"hint":   types.StringValue(ci.Hint),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		checkItemObjs[i] = obj
	}

	checkItemsList, diags := types.ListValue(types.ObjectType{AttrTypes: checkItemAttrTypes}, checkItemObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.CheckItems = checkItemsList

	fieldElems := make(map[string]attr.Value, len(ss.Fields))
	for k, v := range ss.Fields {
		if s, ok := v.(string); ok {
			fieldElems[k] = types.StringValue(s)
		}
	}
	fieldsMap, diags := types.MapValue(types.StringType, fieldElems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Fields = fieldsMap

	if ss.ReplicationJSON != nil {
		state.ReplicationJSON = types.StringValue(*ss.ReplicationJSON)
	} else {
		state.ReplicationJSON = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
