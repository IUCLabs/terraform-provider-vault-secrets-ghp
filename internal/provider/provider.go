package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ghpProvider{}

type ghpProvider struct {
	version string
}

type ghpProviderModel struct {
	Address types.String `tfsdk:"address"`
	Token   types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ghpProvider{version: version}
	}
}

func (p *ghpProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	// The provider's source address is IUCLabs/vault-secrets-ghp,
	// but its resources are prefixed "ghp_" (declare it in required_providers
	// under the local name "ghp"), mirroring how google-beta ships google_*.
	resp.TypeName = "ghp"
	resp.Version = p.version
}

func (p *ghpProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages the ghp secrets engine (vault-plugin-secrets-ghp) on HashiCorp Vault or OpenBao: its mount, connection config and roles. The plugin must already be registered in the plugin catalog.",
		MarkdownDescription: "Manages the [ghp](https://github.com/goodtune/ghp) secrets engine (vault-plugin-secrets-ghp) on HashiCorp Vault or OpenBao: its mount, connection config and roles. The plugin must already be registered in the plugin catalog.",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "Address of the Vault/OpenBao server. Falls back to the VAULT_ADDR environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "Token used to authenticate to Vault/OpenBao. Falls back to the VAULT_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *ghpProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ghpProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := os.Getenv("VAULT_ADDR")
	if !config.Address.IsNull() && config.Address.ValueString() != "" {
		address = config.Address.ValueString()
	}

	token := os.Getenv("VAULT_TOKEN")
	if !config.Token.IsNull() && config.Token.ValueString() != "" {
		token = config.Token.ValueString()
	}

	if address == "" {
		resp.Diagnostics.AddError(
			"missing Vault address",
			"Set the provider \"address\" attribute or the VAULT_ADDR environment variable.",
		)
		return
	}

	client, err := newVaultClient(address, token)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Vault client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ghpProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretBackendResource,
		NewSecretRoleResource,
	}
}

func (p *ghpProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
