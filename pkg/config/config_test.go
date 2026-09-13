package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	prof := Profile{
		Name:           "production",
		KeyID:          "KEY999",
		TeamID:         "TEAM888",
		ClientID:       "SEARCHADS.client777",
		OrgID:          "123456",
		PrivateKeyPath: "/tmp/key.p8",
		BypassKeychain: true,
	}

	err := SaveProfile(prof, true, configPath)
	if err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.ActiveProfile != "production" {
		t.Errorf("expected active profile production, got %s", cfg.ActiveProfile)
	}

	savedProf, exists := cfg.Profiles["production"]
	if !exists {
		t.Fatalf("expected profile 'production' to exist")
	}

	if savedProf.KeyID != "KEY999" {
		t.Errorf("expected KeyID KEY999, got %s", savedProf.KeyID)
	}
	if savedProf.OrgID != "123456" {
		t.Errorf("expected OrgID 123456, got %s", savedProf.OrgID)
	}

	// Test GetActiveProfile
	loadedProf, err := GetActiveProfile("", configPath)
	if err != nil {
		t.Fatalf("GetActiveProfile failed: %v", err)
	}

	if loadedProf.Name != "production" {
		t.Errorf("expected loaded profile name production, got %s", loadedProf.Name)
	}
}
