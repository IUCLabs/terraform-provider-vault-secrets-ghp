package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBackendConfigData(t *testing.T) {
	r := &secretBackendResource{}

	m := secretBackendModel{
		BaseURL:               types.StringValue("https://ghp.example.com"),
		AdminToken:            types.StringValue("ghpsvc_secret"),
		TLSSkipVerify:         types.BoolValue(true),
		RequestTimeoutSeconds: types.Int64Value(15),
		CACert:                types.StringNull(),
	}
	data := r.configData(m)
	if data["base_url"] != "https://ghp.example.com" {
		t.Errorf("base_url = %v", data["base_url"])
	}
	if data["admin_token"] != "ghpsvc_secret" {
		t.Errorf("admin_token = %v", data["admin_token"])
	}
	if data["tls_skip_verify"] != true {
		t.Errorf("tls_skip_verify = %v, want true", data["tls_skip_verify"])
	}
	if data["request_timeout_seconds"] != int64(15) {
		t.Errorf("request_timeout_seconds = %v, want 15", data["request_timeout_seconds"])
	}
	if _, ok := data["ca_cert"]; ok {
		t.Errorf("ca_cert should be omitted when null")
	}

	m.CACert = types.StringValue("-----BEGIN CERTIFICATE-----")
	data = r.configData(m)
	if data["ca_cert"] != "-----BEGIN CERTIFICATE-----" {
		t.Errorf("ca_cert = %v", data["ca_cert"])
	}
}
