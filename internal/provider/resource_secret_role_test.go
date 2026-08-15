package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func listOf(t *testing.T, vals ...string) types.List {
	t.Helper()
	l, diags := types.ListValueFrom(context.Background(), types.StringType, vals)
	if diags.HasError() {
		t.Fatalf("building list: %v", diags)
	}
	return l
}

func TestRoleDataOmitsUnsetFields(t *testing.T) {
	m := secretRoleModel{
		TokenType:      types.StringValue("agent"),
		InstallationID: types.Int64Value(42),
		Scopes:         listOf(t, "contents:read", "pull_requests:write"),
		// AppRecordID, Repositories, SessionPrefix, TTL, MaxTTL left null.
		AppRecordID:   types.StringNull(),
		Repositories:  types.ListNull(types.StringType),
		SessionPrefix: types.StringNull(),
		TTL:           types.Int64Null(),
		MaxTTL:        types.Int64Null(),
	}
	data, diags := roleData(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("roleData: %v", diags)
	}
	if data["token_type"] != "agent" {
		t.Errorf("token_type = %v, want agent", data["token_type"])
	}
	if data["installation_id"] != int64(42) {
		t.Errorf("installation_id = %v, want 42", data["installation_id"])
	}
	if _, ok := data["app_record_id"]; ok {
		t.Errorf("app_record_id should be omitted when null, got %v", data["app_record_id"])
	}
	if _, ok := data["repositories"]; ok {
		t.Errorf("repositories should be omitted when null")
	}
	if _, ok := data["ttl"]; ok {
		t.Errorf("ttl should be omitted when null")
	}
	scopes, ok := data["scopes"].([]string)
	if !ok || len(scopes) != 2 || scopes[0] != "contents:read" {
		t.Errorf("scopes = %v, want [contents:read pull_requests:write]", data["scopes"])
	}
}

func TestApplyRoleDataMapsEngineResponse(t *testing.T) {
	// Shape mirrors what the engine's pathRoleRead returns via the Vault API.
	role := map[string]interface{}{
		"token_type":      "agent",
		"installation_id": json.Number("77"),
		"app_record_id":   "3c2d1e00-0000-4000-8000-000000000000",
		"repositories":    []interface{}{"org/repo"},
		"scopes":          []interface{}{"contents:read"},
		"session_prefix":  "vault",
		"ttl":             json.Number("3600"),
		"max_ttl":         json.Number("86400"),
	}
	var m secretRoleModel
	m.TTL = types.Int64Null()
	m.MaxTTL = types.Int64Null()
	if diags := applyRoleData(&m, role); diags.HasError() {
		t.Fatalf("applyRoleData: %v", diags)
	}
	if m.TokenType.ValueString() != "agent" {
		t.Errorf("token_type = %q, want agent", m.TokenType.ValueString())
	}
	if m.InstallationID.ValueInt64() != 77 {
		t.Errorf("installation_id = %d, want 77", m.InstallationID.ValueInt64())
	}
	if m.AppRecordID.ValueString() != "3c2d1e00-0000-4000-8000-000000000000" {
		t.Errorf("app_record_id = %q", m.AppRecordID.ValueString())
	}
	if m.SessionPrefix.ValueString() != "vault" {
		t.Errorf("session_prefix = %q, want vault", m.SessionPrefix.ValueString())
	}
	if m.TTL.ValueInt64() != 3600 || m.MaxTTL.ValueInt64() != 86400 {
		t.Errorf("ttl/max_ttl = %d/%d, want 3600/86400", m.TTL.ValueInt64(), m.MaxTTL.ValueInt64())
	}
	var scopes []string
	m.Scopes.ElementsAs(context.Background(), &scopes, false)
	if len(scopes) != 1 || scopes[0] != "contents:read" {
		t.Errorf("scopes = %v, want [contents:read]", scopes)
	}
}

func TestApplyRoleDataProxyLeavesInstallationNull(t *testing.T) {
	// A proxy role has no installation; the engine returns 0, which must not
	// clobber the null model value into a spurious 0.
	role := map[string]interface{}{
		"token_type":      "proxy",
		"installation_id": json.Number("0"),
		"app_record_id":   "",
		"repositories":    []interface{}{},
		"scopes":          []interface{}{},
		"session_prefix":  "vault",
		"ttl":             json.Number("0"),
		"max_ttl":         json.Number("0"),
	}
	m := secretRoleModel{
		InstallationID: types.Int64Null(),
		AppRecordID:    types.StringNull(),
		TTL:            types.Int64Null(),
		MaxTTL:         types.Int64Null(),
	}
	if diags := applyRoleData(&m, role); diags.HasError() {
		t.Fatalf("applyRoleData: %v", diags)
	}
	if !m.InstallationID.IsNull() {
		t.Errorf("installation_id = %v, want null for proxy", m.InstallationID)
	}
	if !m.AppRecordID.IsNull() {
		t.Errorf("app_record_id = %v, want null", m.AppRecordID)
	}
	if m.TokenType.ValueString() != "proxy" {
		t.Errorf("token_type = %q, want proxy", m.TokenType.ValueString())
	}
}
