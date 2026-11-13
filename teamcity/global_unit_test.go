package teamcity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"terraform-provider-teamcity/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakeGlobalSettings stands in for TeamCity's /app/rest/server/globalSettings: every response
// reports useEncryption, and the requests it receives are recorded.
type fakeGlobalSettings struct {
	useEncryption bool

	mu   sync.Mutex
	gets int
	puts []map[string]any
}

func (f *fakeGlobalSettings) start(t *testing.T) (*globalResource, resource.SchemaResponse) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case r.URL.Path != "/app/rest/server/globalSettings":
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		case r.Method == http.MethodGet:
			f.gets++
		case r.Method == http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("invalid PUT body: %v", err)
			}
			f.puts = append(f.puts, body)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"useEncryption":%t}`, f.useEncryption)
	}))
	t.Cleanup(server.Close)

	c := client.NewClient(server.URL, "token", "", "", 0)
	r := &globalResource{client: &c}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	return r, schemaResp
}

var testKey = &EncryptionModel{Key: types.StringValue("MTIzNDU2Nzg5MDEyMzQ1Ng==")}

// An imported server that uses a key of its own has no key in the state to carry over. Read must not
// crash on that, and must leave the encryption block out, as a configuration without one does.
func TestGlobalSettingsRead_ImportedServerWithItsOwnKey(t *testing.T) {
	fake := &fakeGlobalSettings{useEncryption: true}
	r, schemaResp := fake.start(t)

	imported := tfsdk.State{Schema: schemaResp.Schema}
	if diags := imported.Set(context.Background(), globalResourceModel{}); diags.HasError() {
		t.Fatalf("failed to set state: %v", diags)
	}
	resp := resource.ReadResponse{State: imported}

	r.Read(context.Background(), resource.ReadRequest{State: imported}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	var got globalResourceModel
	if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("failed reading resulting state: %v", diags)
	}
	if got.Encryption != nil {
		t.Errorf("got an encryption block in the state, want none")
	}
}

// Without an encryption block, an apply keeps the server's own key instead of turning it off.
func TestGlobalSettingsUpdate_KeepsTheServerKeyWithoutABlock(t *testing.T) {
	fake := &fakeGlobalSettings{useEncryption: true}
	r, schemaResp := fake.start(t)

	model := globalResourceModel{
		ArtifactDirectories:            types.StringValue("system/artifacts"),
		RootUrl:                        types.StringValue("https://teamcity.example.com"),
		MaxArtifactSize:                types.Int64Value(1073741824),
		MaxArtifactNumber:              types.Int64Value(-1),
		DefaultExecutionTimeout:        types.Int64Value(0),
		DefaultVCSCheckInterval:        types.Int64Value(60),
		EnforceDefaultVCSCheckInterval: types.BoolValue(false),
		DefaultQuietPeriod:             types.Int64Value(1),
		ArtifactIsolation:              &ArtifactIsolationModel{Enabled: types.BoolValue(false), ArtifactsUrl: types.StringValue("")},
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to set plan: %v", diags)
	}
	prior := tfsdk.State{Schema: schemaResp.Schema}
	if diags := prior.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to set state: %v", diags)
	}
	resp := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: prior}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if len(fake.puts) != 1 {
		t.Fatalf("got %d PUT requests, want 1", len(fake.puts))
	}
	if got := fake.puts[0]["useEncryption"]; got != true {
		t.Errorf("sent useEncryption=%v, want true", got)
	}
	if _, sent := fake.puts[0]["encryptionKey"]; sent {
		t.Errorf("sent an encryptionKey, want none")
	}
	var got globalResourceModel
	if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("failed reading resulting state: %v", diags)
	}
	if got.Encryption != nil {
		t.Errorf("got an encryption block in the state, want none")
	}
}

func TestEncryptionToSet(t *testing.T) {
	tests := []struct {
		name              string
		configured, prior *EncryptionModel
		serverEncrypts    bool
		wantUse           bool
		wantKey           string
		wantGets          int
	}{
		{name: "a configured block sets its key", configured: testKey, prior: testKey, serverEncrypts: false, wantUse: true, wantKey: testKey.Key.ValueString()},
		{name: "a removed block turns the custom key off", configured: nil, prior: testKey, serverEncrypts: true, wantUse: false},
		{name: "no block keeps the server's custom key", configured: nil, prior: nil, serverEncrypts: true, wantUse: true, wantGets: 1},
		{name: "no block keeps the server's default encryption", configured: nil, prior: nil, serverEncrypts: false, wantUse: false, wantGets: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGlobalSettings{useEncryption: tt.serverEncrypts}
			r, _ := fake.start(t)

			gotUse, gotKey, err := r.encryptionToSet(tt.configured, tt.prior)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotUse != tt.wantUse || gotKey != tt.wantKey {
				t.Errorf("got useEncryption=%v key=%q, want useEncryption=%v key=%q", gotUse, gotKey, tt.wantUse, tt.wantKey)
			}
			if fake.gets != tt.wantGets {
				t.Errorf("got %d GET requests, want %d", fake.gets, tt.wantGets)
			}
		})
	}
}

// The key never comes back from TeamCity, so the state keeps the configured one, and holds no block
// at all when the configuration has none.
func TestConfiguredEncryption(t *testing.T) {
	serverBlock := &EncryptionModel{}

	tests := []struct {
		name               string
		server, configured *EncryptionModel
		wantKey            *string
	}{
		{name: "server key, configured block", server: serverBlock, configured: testKey, wantKey: strPtr(testKey.Key.ValueString())},
		{name: "server key, no configured block", server: serverBlock, configured: nil},
		{name: "default encryption, configured block", server: nil, configured: testKey},
		{name: "default encryption, no configured block", server: nil, configured: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := configuredEncryption(tt.server, tt.configured)

			switch {
			case tt.wantKey == nil && got != nil:
				t.Errorf("got an encryption block, want none")
			case tt.wantKey != nil && got == nil:
				t.Errorf("got no encryption block, want one with the configured key")
			case tt.wantKey != nil && got.Key.ValueString() != *tt.wantKey:
				t.Errorf("got key %q, want the configured key", got.Key.ValueString())
			}
		})
	}
}
