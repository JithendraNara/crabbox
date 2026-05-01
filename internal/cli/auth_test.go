package cli

import (
	"path/filepath"
	"testing"
)

func TestWriteBrokerLoginStoresTokenInUserConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("CRABBOX_CONFIG", "")
	t.Setenv("CRABBOX_COORDINATOR", "")
	t.Setenv("CRABBOX_COORDINATOR_TOKEN", "")
	t.Setenv("CRABBOX_PROVIDER", "")

	path, cfg, err := writeBrokerLogin("https://crabbox.example.test", "secret", "aws")
	if err != nil {
		t.Fatal(err)
	}
	if path != userConfigPath() {
		t.Fatalf("path=%q want %q", path, userConfigPath())
	}
	if cfg.Coordinator != "https://crabbox.example.test" || cfg.CoordToken != "secret" || cfg.Provider != "aws" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}
