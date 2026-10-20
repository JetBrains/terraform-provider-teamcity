package teamcity

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"terraform-provider-teamcity/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func newTestBCSettingsResource(serverURL string) (*bcSettingsResource, resource.SchemaResponse) {
	c := client.NewClient(serverURL, "token", "", "", 0)
	r := &bcSettingsResource{client: &c}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	return r, schemaResp
}

func makeBCSettingsPlan(t *testing.T, schemaResp resource.SchemaResponse, model bcSettingsResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to set plan: %v", diags)
	}
	return plan
}

func makeBCSettingsState(t *testing.T, schemaResp resource.SchemaResponse, model bcSettingsResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to set state: %v", diags)
	}
	return state
}

func TestReadSettings_AllPresent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberCounter":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("42"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberPattern":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("1.0.%build.counter%"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/artifactRules":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("+:target/*.jar"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token", "", "", 0)
	r := &bcSettingsResource{client: &c}

	model := bcSettingsResourceModel{
		Id:                   types.StringValue("Test_Bc"),
		BuildConfigurationId: types.StringValue("Test_Bc"),
		BuildNumberCounter:   types.Int64Unknown(),
		BuildNumberPattern:   types.StringUnknown(),
		ArtifactRules:        types.StringValue("+:target/*.jar"),
	}
	var diags diag.Diagnostics

	found, ok := r.readSettings("Test_Bc", &model, &diags)
	if !ok || diags.HasError() {
		t.Fatalf("expected ok without errors, got ok=%v, diags=%v", ok, diags)
	}
	if !found {
		t.Fatalf("expected found=true")
	}
	if model.BuildNumberCounter.IsUnknown() || model.BuildNumberCounter.ValueInt64() != 42 {
		t.Errorf("expected BuildNumberCounter=42, got %v", model.BuildNumberCounter)
	}
	if model.BuildNumberPattern.IsUnknown() || model.BuildNumberPattern.ValueString() != "1.0.%build.counter%" {
		t.Errorf("expected BuildNumberPattern='1.0.%%build.counter%%', got %v", model.BuildNumberPattern)
	}
	if model.ArtifactRules.IsUnknown() || model.ArtifactRules.ValueString() != "+:target/*.jar" {
		t.Errorf("expected ArtifactRules='+:target/*.jar', got %v", model.ArtifactRules)
	}
}

func TestReadSettings_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token", "", "", 0)
	r := &bcSettingsResource{client: &c}

	model := bcSettingsResourceModel{
		Id:                   types.StringValue("Missing_Bc"),
		BuildConfigurationId: types.StringValue("Missing_Bc"),
	}
	var diags diag.Diagnostics

	found, ok := r.readSettings("Missing_Bc", &model, &diags)
	if !ok || diags.HasError() {
		t.Fatalf("expected ok without errors, got ok=%v, diags=%v", ok, diags)
	}
	if found {
		t.Errorf("expected found=false when all settings return 404")
	}
}

func TestReadSettings_InvalidCounterReturnsDiagnostic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/rest/buildTypes/id:Bad_Bc/settings/buildNumberCounter":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-an-int"))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		}
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token", "", "", 0)
	r := &bcSettingsResource{client: &c}

	model := bcSettingsResourceModel{
		Id:                   types.StringValue("Bad_Bc"),
		BuildConfigurationId: types.StringValue("Bad_Bc"),
	}
	var diags diag.Diagnostics

	_, ok := r.readSettings("Bad_Bc", &model, &diags)
	if ok || !diags.HasError() {
		t.Fatalf("expected error diagnostic on non-integer buildNumberCounter, got ok=%v, diags=%v", ok, diags)
	}
}

func TestReadSettings_ServerErrorReturnsDiagnostic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token", "", "", 0)
	r := &bcSettingsResource{client: &c}

	model := bcSettingsResourceModel{
		Id:                   types.StringValue("Err_Bc"),
		BuildConfigurationId: types.StringValue("Err_Bc"),
	}
	var diags diag.Diagnostics

	found, ok := r.readSettings("Err_Bc", &model, &diags)
	if ok || found || !diags.HasError() {
		t.Fatalf("expected found=false, ok=false, diags.HasError()=true; got found=%v, ok=%v, diags=%v", found, ok, diags)
	}
}

// TestBCSettingsCreate_SkipsUnknownAndReadsBack verifies that when optional+computed
// attributes are Unknown in the plan (omitted in HCL), Create does NOT send PUT
// requests for the omitted attributes and populates concrete values from TeamCity.
func TestBCSettingsCreate_SkipsUnknownAndReadsBack(t *testing.T) {
	var mu sync.Mutex
	putCalls := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			putCalls[r.URL.Path] = string(body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
			return
		}

		switch r.URL.Path {
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberCounter":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("77"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberPattern":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("existing-%build.counter%"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/artifactRules":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("+:new/*.jar"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	r, schemaResp := newTestBCSettingsResource(server.URL)

	planModel := bcSettingsResourceModel{
		Id:                   types.StringUnknown(),
		BuildConfigurationId: types.StringValue("Test_Bc"),
		BuildNumberCounter:   types.Int64Unknown(),
		BuildNumberPattern:   types.StringUnknown(),
		ArtifactRules:        types.StringValue("+:new/*.jar"),
	}

	req := resource.CreateRequest{
		Plan: makeBCSettingsPlan(t, schemaResp, planModel),
	}
	resp := resource.CreateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	if len(putCalls) != 1 {
		t.Fatalf("expected exactly 1 PUT call (only artifactRules), got %d: %v", len(putCalls), putCalls)
	}
	if got, ok := putCalls["/app/rest/buildTypes/id:Test_Bc/settings/artifactRules"]; !ok || got != "+:new/*.jar" {
		t.Errorf("unexpected artifactRules PUT call: %v", putCalls)
	}

	var gotState bcSettingsResourceModel
	if diags := resp.State.Get(context.Background(), &gotState); diags.HasError() {
		t.Fatalf("failed reading resulting state: %v", diags)
	}
	if gotState.Id.ValueString() != "Test_Bc" {
		t.Errorf("expected Id='Test_Bc', got %v", gotState.Id)
	}
	if gotState.BuildNumberCounter.IsUnknown() || gotState.BuildNumberCounter.ValueInt64() != 77 {
		t.Errorf("expected preserved BuildNumberCounter=77, got %v", gotState.BuildNumberCounter)
	}
	if gotState.BuildNumberPattern.IsUnknown() || gotState.BuildNumberPattern.ValueString() != "existing-%build.counter%" {
		t.Errorf("expected preserved BuildNumberPattern='existing-%%build.counter%%', got %v", gotState.BuildNumberPattern)
	}
	if gotState.ArtifactRules.IsUnknown() || gotState.ArtifactRules.ValueString() != "+:new/*.jar" {
		t.Errorf("expected ArtifactRules='+:new/*.jar', got %v", gotState.ArtifactRules)
	}
}

// TestBCSettingsUpdate_OnlyUpdatesChangedKnownAttributes verifies that Update only
// sends PUT requests for attributes that are concrete (non-null, non-unknown) and
// changed relative to prior state, then refreshes state from TeamCity.
func TestBCSettingsUpdate_OnlyUpdatesChangedKnownAttributes(t *testing.T) {
	var mu sync.Mutex
	putCalls := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			putCalls[r.URL.Path] = string(body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
			return
		}

		switch r.URL.Path {
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberCounter":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("77"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberPattern":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("updated-%build.counter%"))
		case "/app/rest/buildTypes/id:Test_Bc/settings/artifactRules":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("+:new/*.jar"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	r, schemaResp := newTestBCSettingsResource(server.URL)

	priorStateModel := bcSettingsResourceModel{
		Id:                   types.StringValue("Test_Bc"),
		BuildConfigurationId: types.StringValue("Test_Bc"),
		BuildNumberCounter:   types.Int64Value(77),
		BuildNumberPattern:   types.StringValue("existing-%build.counter%"),
		ArtifactRules:        types.StringValue("+:new/*.jar"),
	}
	planModel := bcSettingsResourceModel{
		Id:                   types.StringValue("Test_Bc"),
		BuildConfigurationId: types.StringValue("Test_Bc"),
		BuildNumberCounter:   types.Int64Unknown(), // e.g. omitted in HCL without prior value or unknown
		BuildNumberPattern:   types.StringValue("updated-%build.counter%"),
		ArtifactRules:        types.StringValue("+:new/*.jar"), // unchanged
	}

	req := resource.UpdateRequest{
		State: makeBCSettingsState(t, schemaResp, priorStateModel),
		Plan:  makeBCSettingsPlan(t, schemaResp, planModel),
	}
	resp := resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	if len(putCalls) != 1 {
		t.Fatalf("expected exactly 1 PUT call (only buildNumberPattern), got %d: %v", len(putCalls), putCalls)
	}
	if got, ok := putCalls["/app/rest/buildTypes/id:Test_Bc/settings/buildNumberPattern"]; !ok || got != "updated-%build.counter%" {
		t.Errorf("unexpected buildNumberPattern PUT call: %v", putCalls)
	}

	var gotState bcSettingsResourceModel
	if diags := resp.State.Get(context.Background(), &gotState); diags.HasError() {
		t.Fatalf("failed reading resulting state: %v", diags)
	}
	if gotState.BuildNumberCounter.IsUnknown() || gotState.BuildNumberCounter.ValueInt64() != 77 {
		t.Errorf("expected BuildNumberCounter=77 in state, got %v", gotState.BuildNumberCounter)
	}
	if gotState.BuildNumberPattern.ValueString() != "updated-%build.counter%" {
		t.Errorf("expected BuildNumberPattern='updated-%%build.counter%%', got %v", gotState.BuildNumberPattern)
	}
}

// TestBCSettingsRead_RemovesResourceWhenAllSettingsNotFound verifies that Read
// removes the resource from Terraform state when the build configuration no longer exists.
func TestBCSettingsRead_RemovesResourceWhenAllSettingsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	r, schemaResp := newTestBCSettingsResource(server.URL)

	priorStateModel := bcSettingsResourceModel{
		Id:                   types.StringValue("Deleted_Bc"),
		BuildConfigurationId: types.StringValue("Deleted_Bc"),
		BuildNumberCounter:   types.Int64Value(1),
		BuildNumberPattern:   types.StringValue("%build.counter%"),
		ArtifactRules:        types.StringValue(""),
	}

	state := makeBCSettingsState(t, schemaResp, priorStateModel)
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.Equal(tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil)) {
		t.Errorf("expected state to be removed (null), got: %v", resp.State.Raw)
	}
}

func TestBCSettingsImportState_SetsIdAndBuildConfigurationId(t *testing.T) {
	r, schemaResp := newTestBCSettingsResource("http://localhost:8111")

	req := resource.ImportStateRequest{ID: "Imported_Bc"}
	resp := resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil),
		},
	}

	r.ImportState(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var gotState bcSettingsResourceModel
	if diags := resp.State.Get(context.Background(), &gotState); diags.HasError() {
		t.Fatalf("failed reading state after ImportState: %v", diags)
	}
	if gotState.Id.ValueString() != "Imported_Bc" {
		t.Errorf("expected Id='Imported_Bc', got %v", gotState.Id)
	}
	if gotState.BuildConfigurationId.ValueString() != "Imported_Bc" {
		t.Errorf("expected BuildConfigurationId='Imported_Bc', got %v", gotState.BuildConfigurationId)
	}
}
