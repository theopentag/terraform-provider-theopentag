package sql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/theopentag/terraform-provider-theopentag/internal/client"
)

var _ datasource.DataSource = &sshKeyDataSource{}

type sshKeyDataSource struct {
	client *client.Client
}

type sshKeyModel struct {
	PublicKey types.String `tfsdk:"public_key"`
}

func NewSSHKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

func (d *sshKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_ssh_key"
}

func (d *sshKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the SSH public key used by SQL for remote restore operations.",
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "Ed25519 public key in OpenSSH format. Null if not yet generated.",
			},
		},
	}
}

func (d *sshKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	key, err := d.client.GetSSHKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SSH key", err.Error())
		return
	}

	state := sshKeyModel{}
	if key != nil && key.PublicKey != nil {
		state.PublicKey = types.StringValue(*key.PublicKey)
	} else {
		state.PublicKey = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
