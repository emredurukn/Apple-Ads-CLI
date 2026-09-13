package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	KeyringService = "asactl"
	DefaultDir     = ".asactl"
	ConfigFile     = "config.yaml"
)

// Profile holds the credentials and settings for an Apple Ads account.
type Profile struct {
	Name           string `json:"name" yaml:"name" mapstructure:"name"`
	KeyID          string `json:"key_id" yaml:"key_id" mapstructure:"key_id"`
	TeamID         string `json:"team_id" yaml:"team_id" mapstructure:"team_id"`
	ClientID       string `json:"client_id" yaml:"client_id" mapstructure:"client_id"`
	OrgID          string `json:"org_id" yaml:"org_id" mapstructure:"org_id"`
	PrivateKeyPath string `json:"private_key_path" yaml:"private_key_path" mapstructure:"private_key_path"`
	BypassKeychain bool   `json:"bypass_keychain" yaml:"bypass_keychain" mapstructure:"bypass_keychain"`
}

// Config holds the multi-profile configuration file structure.
type Config struct {
	ActiveProfile string             `json:"active_profile" yaml:"active_profile" mapstructure:"active_profile"`
	Profiles      map[string]Profile `json:"profiles" yaml:"profiles" mapstructure:"profiles"`
}

// GetConfigDir returns the ~/.apple-ads path.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home directory: %w", err)
	}
	return filepath.Join(home, DefaultDir), nil
}

// EnsureConfigDir creates ~/.apple-ads if it does not exist.
func EnsureConfigDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}
	return dir, nil
}

// GetConfigFilePath returns the full path to config.yaml.
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// LoadConfig reads the config file from disk.
func LoadConfig(customPath string) (*Config, error) {
	path := customPath
	if path == "" {
		var err error
		path, err = GetConfigFilePath()
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		ActiveProfile: "default",
		Profiles:      make(map[string]Profile),
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, nil
}

// SaveConfig writes the configuration back to disk.
func SaveConfig(cfg *Config, customPath string) error {
	path := customPath
	if path == "" {
		dir, err := EnsureConfigDir()
		if err != nil {
			return err
		}
		path = filepath.Join(dir, ConfigFile)
	} else {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0700)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveProfile saves or updates a profile in the configuration and optionally keychain.
func SaveProfile(p Profile, makeActive bool, customPath string) error {
	cfg, err := LoadConfig(customPath)
	if err != nil {
		return err
	}

	if p.Name == "" {
		p.Name = "default"
	}

	cfg.Profiles[p.Name] = p
	if makeActive || cfg.ActiveProfile == "" {
		cfg.ActiveProfile = p.Name
	}

	return SaveConfig(cfg, customPath)
}

// GetActiveProfile retrieves the active profile taking into account flag, env, and config.
func GetActiveProfile(profileFlag string, customPath string) (*Profile, error) {
	cfg, err := LoadConfig(customPath)
	if err != nil {
		return nil, err
	}

	targetProfile := profileFlag
	if targetProfile == "" {
		targetProfile = viper.GetString("profile")
	}
	if targetProfile == "" {
		targetProfile = os.Getenv("APPLE_ADS_PROFILE")
	}
	if targetProfile == "" {
		targetProfile = cfg.ActiveProfile
	}
	if targetProfile == "" {
		targetProfile = "default"
	}

	// Environment variable overrides (supports ASACTL_ and APPLE_ADS_)
	getEnv := func(key string) string {
		val := os.Getenv("ASACTL_" + key)
		if val == "" {
			val = os.Getenv("APPLE_ADS_" + key)
		}
		return val
	}

	envKeyID := getEnv("KEY_ID")
	envTeamID := getEnv("TEAM_ID")
	envClientID := getEnv("CLIENT_ID")
	envOrgID := getEnv("ORG_ID")
	envPrivateKeyPath := getEnv("PRIVATE_KEY_PATH")

	if envKeyID != "" && envTeamID != "" && envClientID != "" && envPrivateKeyPath != "" {
		return &Profile{
			Name:           "env",
			KeyID:          envKeyID,
			TeamID:         envTeamID,
			ClientID:       envClientID,
			OrgID:          envOrgID,
			PrivateKeyPath: envPrivateKeyPath,
			BypassKeychain: true,
		}, nil
	}

	prof, exists := cfg.Profiles[targetProfile]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found. Run 'asactl auth login' to configure", targetProfile)
	}

	return &prof, nil
}

// StorePrivateKeyInKeyring stores the raw private key in the OS keyring.
func StorePrivateKeyInKeyring(profileName string, privateKeyPEM string) error {
	return keyring.Set(KeyringService, profileName, privateKeyPEM)
}

// GetPrivateKeyFromKeyring retrieves the raw private key from the OS keyring.
func GetPrivateKeyFromKeyring(profileName string) (string, error) {
	return keyring.Get(KeyringService, profileName)
}

// DeletePrivateKeyFromKeyring removes the private key from the OS keyring.
func DeletePrivateKeyFromKeyring(profileName string) error {
	return keyring.Delete(KeyringService, profileName)
}
