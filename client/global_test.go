package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TeamCity rejects an empty encryptionKey while useEncryption is true, so an empty key must not be
// sent at all; TeamCity then keeps its current key.
func boolPtr(v bool) *bool {
	return &v
}

func TestSetGlobalSettingsEncryptionKey(t *testing.T) {
	tests := []struct {
		name              string
		useEncryption     *bool
		key               string
		wantUseEncryption *bool
		wantKey           bool
	}{
		{name: "unmanaged encryption omits both fields", useEncryption: nil, key: "", wantUseEncryption: nil, wantKey: false},
		{name: "disabling encryption sends useEncryption=false without key", useEncryption: boolPtr(false), key: "", wantUseEncryption: boolPtr(false), wantKey: false},
		{name: "an empty key is left out", useEncryption: boolPtr(true), key: "", wantUseEncryption: boolPtr(true), wantKey: false},
		{name: "a key is sent", useEncryption: boolPtr(true), key: "MTIzNDU2Nzg5MDEyMzQ1Ng==", wantUseEncryption: boolPtr(true), wantKey: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sent map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/app/rest/server/globalSettings" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(body, &sent); err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"useEncryption":true}`))
			}))
			defer server.Close()

			httpClient := NewClient(server.URL, "token", "", "", 0)
			if _, err := httpClient.SetGlobalSettings(GlobalSettings{UseEncryption: tt.useEncryption, EncryptionKey: tt.key}); err != nil {
				t.Fatal(err)
			}

			gotUse, usePresent := sent["useEncryption"]
			if tt.wantUseEncryption == nil {
				if usePresent {
					t.Fatalf("useEncryption sent: %v, want omitted", gotUse)
				}
			} else {
				if !usePresent || gotUse != *tt.wantUseEncryption {
					t.Fatalf("sent useEncryption=%v (present=%v), want %v", gotUse, usePresent, *tt.wantUseEncryption)
				}
			}

			got, present := sent["encryptionKey"]
			if present != tt.wantKey {
				t.Fatalf("encryptionKey sent: %v, want sent: %v", present, tt.wantKey)
			}
			if tt.wantKey && got != tt.key {
				t.Errorf("sent encryptionKey %q, want %q", got, tt.key)
			}
		})
	}
}
