package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
	sqlmod "github.com/theopentag/terraform-provider-theopentag/internal/modules/sql"
)

var _ provider.Provider = &sqlProvider{}

type sqlProvider struct{}

type sqlProviderModel struct {
	Host               types.String `tfsdk:"host"`
	APIKey             types.String `tfsdk:"api_key"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func New() provider.Provider {
	return &sqlProvider{}
}

func (p *sqlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "theopentag"
}

func (p *sqlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "SQL API base URL (e.g. https://cloud.example.com). Can also be set via PLATFORM_API_HOST env var.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SQL API key (bmk_...). Can also be set via PLATFORM_API_KEY env var.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification.",
			},
		},
	}
}

func (p *sqlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config sqlProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("PLATFORM_API_HOST")
	if !config.Host.IsNull() && !config.Host.IsUnknown() {
		host = config.Host.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddError("Missing host", "Set the host attribute or PLATFORM_API_HOST environment variable.")
		return
	}

	apiKey := os.Getenv("PLATFORM_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing api_key", "Set the api_key attribute or PLATFORM_API_KEY environment variable.")
		return
	}

	insecure := false
	if !config.InsecureSkipVerify.IsNull() && !config.InsecureSkipVerify.IsUnknown() {
		insecure = config.InsecureSkipVerify.ValueBool()
	}

	c := client.New(host, apiKey, insecure)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *sqlProvider) Resources(_ context.Context) []func() resource.Resource {
	var out []func() resource.Resource
	out = append(out, sqlmod.Resources()...)
	// out = append(out, computemod.Resources()...)  // future: compute module
	// out = append(out, iammod.Resources()...)       // future: iam module
	return out
}

func (p *sqlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var out []func() datasource.DataSource
	out = append(out, sqlmod.DataSources()...)
	// out = append(out, computemod.DataSources()...)  // future: compute module
	// out = append(out, iammod.DataSources()...)       // future: iam module
	return out
}
