package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigTimeoutPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, fileValue, envValue string
		flagValue                 time.Duration
		want                      time.Duration
	}{
		{"default", "", "", 0, 30 * time.Second},
		{"file", "10s", "", 0, 10 * time.Second},
		{"environment", "10s", "20s", 0, 20 * time.Second},
		{"flag", "10s", "20s", time.Second, time.Second},
		{"environment overrides invalid file value", "invalid", "20s", 0, 20 * time.Second},
		{"flag overrides invalid environment value", "10s", "invalid", time.Second, time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TASKAIO_TIMEOUT", tc.envValue)
			t.Setenv("TASKAIO_OUTPUT", "")
			path := filepath.Join(t.TempDir(), "config.yaml")
			if tc.fileValue != "" {
				if err := os.WriteFile(path, []byte("timeout: "+tc.fileValue+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := LoadConfig(path, "", "", "", tc.flagValue)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Timeout != tc.want {
				t.Fatalf("timeout = %s, want %s", cfg.Timeout, tc.want)
			}
		})
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"invalid", "0s", "-1s"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("TASKAIO_TIMEOUT", value)
			if _, err := LoadConfig(filepath.Join(t.TempDir(), "config.yaml"), "", "", "", 0); err == nil {
				t.Fatalf("expected error for timeout %q", value)
			}
		})
	}
}

func TestSaveConfigPermissionsAndRoundTrip(t *testing.T) {
	t.Setenv("TASKAIO_BASE_URL", "")
	t.Setenv("TASKAIO_TOKEN", "")
	t.Setenv("TASKAIO_OUTPUT", "")
	t.Setenv("TASKAIO_TIMEOUT", "")
	path := filepath.Join(t.TempDir(), "taskaio", "config.yaml")
	want := &Config{BaseURL: "https://taskaio.example", Token: "test-token", Output: "table", Timeout: time.Minute}
	for _, overwrite := range []bool{false, true} {
		if overwrite {
			if err := os.Chmod(path, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := SaveConfig(want, path); err != nil {
			t.Fatal(err)
		}
		for target, permission := range map[string]os.FileMode{filepath.Dir(path): 0700, path: 0600} {
			info, err := os.Stat(target)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != permission {
				t.Fatalf("%s permissions = %o, want %o", target, info.Mode().Perm(), permission)
			}
		}
		got, err := LoadConfig(path, "", "", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if got.Token != want.Token || got.BaseURL != want.BaseURL || got.Output != want.Output || got.Timeout != want.Timeout {
			t.Fatal("saved configuration did not round trip")
		}
	}
}
