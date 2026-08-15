package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &secretRoleResource{}
	_ resource.ResourceWithImportState    = &secretRoleResource{}
	_ resource.ResourceWithValidateConfig = &secretRoleResource{}
)

const (
	tokenTypeAgent = "agent"
	tokenTypeProxy = "proxy"
)

type secretRoleResource struct {
	client *vaultClient
}

type secretRoleModel struct {
	Backend        types.String `tfsdk:"backend"`
	Name           types.String `tfsdk:"name"`
	TokenType      types.String `tfsdk:"token_type"`
	InstallationID types.Int64  `tfsdk:"installation_id"`
	AppRecordID    types.String `tfsdk:"app_record_id"`
	Repositories   types.List   `tfsdk:"repositories"`
	Scopes         types.List   `tfsdk:"scopes"`
	SessionPrefix  types.String `tfsdk:"session_prefix"`
	TTL            types.Int64  `tfsdk:"ttl"`
	MaxTTL         types.Int64  `tfsdk:"max_ttl"`
}

func NewSecretRoleResource() resource.Resource {
	return &secretRoleResource{}
}

func (r *secretRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_role"
}

func (r *secretRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a role on the ghp secrets engine that mints short-lived scoped ghp tokens.",
		Attributes: map[string]schema.Attribute{
			"backend": schema.StringAttribute{
				Description: "Mount path of the ghp secrets engine.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the role.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token_type": schema.StringAttribute{
				Description: "ghp token type to mint: \"agent\" (default) or \"proxy\".",
				Optional:    true,
				Computed:    true,
			},
			"installation_id": schema.Int64Attribute{
				Description: "ghp App installation id the minted agent token is bound to (required when token_type is \"agent\").",
				Optional:    true,
			},
			"app_record_id": schema.StringAttribute{
				Description: "Optional ghp App record id (UUID) to pin agent tokens to; empty selects ghp's default app.",
				Optional:    true,
			},
			"repositories": schema.ListAttribute{
				Description: "Optional repositories the minted token is restricted to; empty is open-scoped (all repositories).",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Optional ghp permission:level scopes (e.g. [\"contents:read\", \"pull_requests:write\"]); empty is open-scoped.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"session_prefix": schema.StringAttribute{
				Description: "Prefix for the ghp session id of each minted token (default \"vault\").",
				Optional:    true,
				Computed:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Default lease TTL in seconds for tokens minted from this role.",
				Optional:    true,
			},
			"max_ttl": schema.Int64Attribute{
				Description: "Maximum lease TTL in seconds for tokens minted from this role.",
				Optional:    true,
			},
		},
	}
}

func (r *secretRoleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config secretRoleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// token_type defaults to "agent" server-side, so a null token_type is an
	// agent role and still requires installation_id.
	isAgent := config.TokenType.IsNull() || config.TokenType.ValueString() == tokenTypeAgent
	if isAgent && config.InstallationID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("installation_id"),
			"installation_id required for agent tokens",
			"token_type is \"agent\" (the default), which requires installation_id to be set.",
		)
	}
}

func (r *secretRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*vaultClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *secretRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretRoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data, diags := roleData(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.write(ctx, rolePath(plan.Backend.ValueString(), plan.Name.ValueString()), data); err != nil {
		resp.Diagnostics.AddError("failed to create ghp role", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readInto(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *secretRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretRoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.read(ctx, rolePath(state.Backend.ValueString(), state.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed to read ghp role", err.Error())
		return
	}
	if role == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(applyRoleData(&state, role)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *secretRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secretRoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data, diags := roleData(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.write(ctx, rolePath(plan.Backend.ValueString(), plan.Name.ValueString()), data); err != nil {
		resp.Diagnostics.AddError("failed to update ghp role", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readInto(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *secretRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretRoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.delete(ctx, rolePath(state.Backend.ValueString(), state.Name.ValueString())); err != nil {
		resp.Diagnostics.AddError("failed to delete ghp role", err.Error())
		return
	}
}

func (r *secretRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	backend, name, ok := splitBackendName(req.ID, "roles")
	if !ok {
		resp.Diagnostics.AddError(
			"invalid import ID",
			fmt.Sprintf("expected \"<backend>/roles/<name>\", got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("backend"), backend)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func (r *secretRoleResource) readInto(ctx context.Context, m *secretRoleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	role, err := r.client.read(ctx, rolePath(m.Backend.ValueString(), m.Name.ValueString()))
	if err != nil {
		diags.AddError("failed to read back ghp role", err.Error())
		return diags
	}
	if role == nil {
		diags.AddError("role missing after write", "the role was not found immediately after being written")
		return diags
	}
	return applyRoleData(m, role)
}

func roleData(ctx context.Context, m secretRoleModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	data := map[string]interface{}{}

	if !m.TokenType.IsNull() && !m.TokenType.IsUnknown() {
		data["token_type"] = m.TokenType.ValueString()
	}
	if !m.InstallationID.IsNull() && !m.InstallationID.IsUnknown() {
		data["installation_id"] = m.InstallationID.ValueInt64()
	}
	if !m.AppRecordID.IsNull() && !m.AppRecordID.IsUnknown() {
		data["app_record_id"] = m.AppRecordID.ValueString()
	}
	if !m.Repositories.IsNull() && !m.Repositories.IsUnknown() {
		var repos []string
		diags.Append(m.Repositories.ElementsAs(ctx, &repos, false)...)
		data["repositories"] = repos
	}
	if !m.Scopes.IsNull() && !m.Scopes.IsUnknown() {
		var scopes []string
		diags.Append(m.Scopes.ElementsAs(ctx, &scopes, false)...)
		data["scopes"] = scopes
	}
	if !m.SessionPrefix.IsNull() && !m.SessionPrefix.IsUnknown() {
		data["session_prefix"] = m.SessionPrefix.ValueString()
	}
	if !m.TTL.IsNull() && !m.TTL.IsUnknown() {
		data["ttl"] = m.TTL.ValueInt64()
	}
	if !m.MaxTTL.IsNull() && !m.MaxTTL.IsUnknown() {
		data["max_ttl"] = m.MaxTTL.ValueInt64()
	}
	return data, diags
}

func applyRoleData(m *secretRoleModel, role map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if v, ok := role["token_type"].(string); ok && v != "" {
		m.TokenType = types.StringValue(v)
	}
	if n, ok := toInt64(role["installation_id"]); ok && n != 0 {
		m.InstallationID = types.Int64Value(n)
	}
	if v, ok := role["app_record_id"].(string); ok && v != "" {
		m.AppRecordID = types.StringValue(v)
	}

	repos := toStringSlice(role["repositories"])
	repoList, repoDiags := types.ListValueFrom(context.Background(), types.StringType, repos)
	diags.Append(repoDiags...)
	m.Repositories = repoList

	scopeVals := toStringSlice(role["scopes"])
	scopeList, scopeDiags := types.ListValueFrom(context.Background(), types.StringType, scopeVals)
	diags.Append(scopeDiags...)
	m.Scopes = scopeList

	if v, ok := role["session_prefix"].(string); ok && v != "" {
		m.SessionPrefix = types.StringValue(v)
	}
	if n, ok := toInt64(role["ttl"]); ok && n != 0 {
		m.TTL = types.Int64Value(n)
	} else if m.TTL.IsUnknown() {
		m.TTL = types.Int64Null()
	}
	if n, ok := toInt64(role["max_ttl"]); ok && n != 0 {
		m.MaxTTL = types.Int64Value(n)
	} else if m.MaxTTL.IsUnknown() {
		m.MaxTTL = types.Int64Null()
	}
	return diags
}
