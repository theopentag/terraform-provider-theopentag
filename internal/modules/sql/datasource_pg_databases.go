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

var _ datasource.DataSource = &pgDatabasesDataSource{}

type pgDatabasesDataSource struct {
	client *client.Client
}

type pgDatabasesModel struct {
	ServerName types.String `tfsdk:"server_name"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	Databases  types.List   `tfsdk:"databases"`
}

var pgDatabaseAttrTypes = map[string]attr.Type{
	"database_name":      types.StringType,
	"owner":              types.StringType,
	"encoding":           types.StringType,
	"collation":          types.StringType,
	"size_bytes":         types.Int64Type,
	"connection_limit":   types.Int64Type,
	"is_template":        types.BoolType,
	"allows_connections": types.BoolType,
}

func NewPGDatabasesDataSource() datasource.DataSource {
	return &pgDatabasesDataSource{}
}

func (d *pgDatabasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_pg_databases"
}

func (d *pgDatabasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Snapshot of PostgreSQL databases on a managed server (refreshed every 30s by the backend).",
		Attributes: map[string]schema.Attribute{
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SQL-managed server.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last snapshot (ISO UTC).",
			},
			"databases": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of databases.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							Computed:    true,
							Description: "Database name.",
						},
						"owner": schema.StringAttribute{
							Computed:    true,
							Description: "Database owner role.",
						},
						"encoding": schema.StringAttribute{
							Computed:    true,
							Description: "Character encoding.",
						},
						"collation": schema.StringAttribute{
							Computed:    true,
							Description: "Collation.",
						},
						"size_bytes": schema.Int64Attribute{
							Computed:    true,
							Description: "Database size in bytes.",
						},
						"connection_limit": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum connections (-1 = unlimited).",
						},
						"is_template": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is a template database.",
						},
						"allows_connections": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether connections are allowed.",
						},
					},
				},
			},
		},
	}
}

func (d *pgDatabasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pgDatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pgDatabasesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.client.GetPGDatabases(ctx, config.ServerName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading pg databases", err.Error())
		return
	}

	objs := make([]attr.Value, len(r.Databases))
	for i, db := range r.Databases {
		obj, diags := types.ObjectValue(pgDatabaseAttrTypes, map[string]attr.Value{
			"database_name":      types.StringValue(db.DatabaseName),
			"owner":              types.StringValue(db.Owner),
			"encoding":           types.StringValue(db.Encoding),
			"collation":          types.StringValue(db.Collation),
			"size_bytes":         types.Int64Value(db.SizeBytes),
			"connection_limit":   types.Int64Value(db.ConnectionLimit),
			"is_template":        types.BoolValue(db.IsTemplate),
			"allows_connections": types.BoolValue(db.AllowsConnections),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: pgDatabaseAttrTypes}, objs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedAt := types.StringNull()
	if r.UpdatedAt != nil {
		updatedAt = types.StringValue(*r.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, pgDatabasesModel{
		ServerName: config.ServerName,
		UpdatedAt:  updatedAt,
		Databases:  list,
	})...)
}
