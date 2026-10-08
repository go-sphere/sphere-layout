package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sphere/sphere-layout/internal/server/dash"
)

func TestNewEmptyConfigProvidesUsableDefaults(t *testing.T) {
	config := NewEmptyConfig()
	if config.Dash.AuthJWT == "" || config.Dash.RefreshJWT == "" || config.API.JWT == "" {
		t.Fatal("generated authentication secrets must be non-empty")
	}
	if config.Dash.AuthJWT == config.Dash.RefreshJWT || config.Dash.AuthJWT == config.API.JWT || config.Dash.RefreshJWT == config.API.JWT {
		t.Fatal("generated authentication secrets must be independent")
	}
	if config.Database.AutoMigrateDrop {
		t.Fatal("destructive database migration must be disabled by default")
	}
	if config.Dash.SeedUser.Username == "" || config.Dash.SeedUser.Password == "" {
		t.Fatal("default seed-user settings must be non-empty")
	}
}

func TestNewConfigValidatesRequiredSecrets(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing dashboard secrets",
			content: `{"api":{"jwt":"api-secret"}}`,
			wantErr: "dash auth_jwt and refresh_jwt must be non-empty",
		},
		{
			name:    "missing API secret",
			content: `{"dash":{"auth_jwt":"auth-secret","refresh_jwt":"refresh-secret"}}`,
			wantErr: "api jwt must be non-empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfig(writeConfig(t, tt.content))
			if err == nil {
				t.Fatalf("NewConfig() error = nil, want %q", tt.wantErr)
			}
			if got := err.Error(); got != tt.wantErr {
				t.Errorf("NewConfig() error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestNewConfigAppliesLogLevelDefault(t *testing.T) {
	config, err := NewConfig(writeConfig(t, `{
		"dash":{"auth_jwt":"auth-secret","refresh_jwt":"refresh-secret"},
		"api":{"jwt":"api-secret"}
	}`))
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if got, want := config.Log.Level, "info"; got != want {
		t.Errorf("Log.Level = %q, want %q", got, want)
	}
}

func TestNewConfigLoadsSeedUserSettings(t *testing.T) {
	t.Run("omitted seed user uses default username and no password", func(t *testing.T) {
		config, err := NewConfig(writeConfig(t, `{
			"dash":{"auth_jwt":"auth-secret","refresh_jwt":"refresh-secret"},
			"api":{"jwt":"api-secret"}
		}`))
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}
		if config.Dash.SeedUser.Username != dash.DefaultSeedUsername {
			t.Errorf("SeedUser.Username = %q, want %q", config.Dash.SeedUser.Username, dash.DefaultSeedUsername)
		}
		if config.Dash.SeedUser.Password != "" {
			t.Errorf("SeedUser.Password = %q, want empty", config.Dash.SeedUser.Password)
		}
	})
	t.Run("explicit seed user is loaded", func(t *testing.T) {
		config, err := NewConfig(writeConfig(t, `{
			"dash":{
				"auth_jwt":"auth-secret",
				"refresh_jwt":"refresh-secret",
				"seed_user":{"username":"from-file","password":"FromFile#1"}
			},
			"api":{"jwt":"api-secret"}
		}`))
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}
		if config.Dash.SeedUser.Username != "from-file" {
			t.Errorf("SeedUser.Username = %q, want from-file", config.Dash.SeedUser.Username)
		}
		if config.Dash.SeedUser.Password != "FromFile#1" {
			t.Errorf("SeedUser.Password = %q, want FromFile#1", config.Dash.SeedUser.Password)
		}
	})
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestNewConfigLoadsHTTPServerOptions(t *testing.T) {
	config, err := NewConfig(writeConfig(t, `{
		"dash":{"auth_jwt":"auth-secret","refresh_jwt":"refresh-secret",
			"http":{"trusted_proxies":["10.0.0.0/8"],"max_body_bytes":1024,"read_timeout_seconds":5}},
		"api":{"jwt":"api-secret","http":{"idle_timeout_seconds":7}},
		"file":{"max_body_bytes":-1}
	}`))
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if got := config.Dash.HTTP.TrustedProxies; len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Errorf("Dash.HTTP.TrustedProxies = %v, want [10.0.0.0/8]", got)
	}
	if config.Dash.HTTP.MaxBodyBytes != 1024 || config.Dash.HTTP.ReadTimeoutSeconds != 5 {
		t.Errorf("Dash.HTTP options = %+v", config.Dash.HTTP.Options)
	}
	if config.API.HTTP.IdleTimeoutSeconds != 7 {
		t.Errorf("API.HTTP.IdleTimeoutSeconds = %d, want 7", config.API.HTTP.IdleTimeoutSeconds)
	}
	if config.File.MaxBodyBytes != -1 {
		t.Errorf("File.MaxBodyBytes = %d, want -1", config.File.MaxBodyBytes)
	}
}

func TestNewConfigRejectsInvalidTrustedProxy(t *testing.T) {
	_, err := NewConfig(writeConfig(t, `{
		"dash":{"auth_jwt":"auth-secret","refresh_jwt":"refresh-secret"},
		"api":{"jwt":"api-secret","http":{"trusted_proxies":["proxy.internal"]}}
	}`))
	if err == nil || !strings.Contains(err.Error(), "api http: trusted_proxies") {
		t.Fatalf("NewConfig() error = %v, want an api http trusted_proxies error", err)
	}
}
