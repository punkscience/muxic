// Package config manages the ~/.muxic/config.json file for storing
// service credentials and OAuth tokens.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// Config is the top-level configuration structure.
type Config struct {
	Spotify SpotifyConfig `json:"spotify,omitempty"`
	YouTube YouTubeConfig `json:"youtube,omitempty"`
}

// SpotifyConfig holds Spotify OAuth credentials and token.
type SpotifyConfig struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	Token        *oauth2.Token `json:"token,omitempty"`
}

// YouTubeConfig holds the YouTube OAuth token. Credentials are read from
// ~/.muxic/youtube_credentials.json (downloaded from Google Cloud Console).
type YouTubeConfig struct {
	Token *oauth2.Token `json:"token,omitempty"`
}

// ConfigPath returns the path to ~/.muxic/config.json.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".muxic", "config.json")
	}
	return filepath.Join(home, ".muxic", "config.json")
}

// Load reads config from disk. Returns an empty Config if the file does not exist.
func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes cfg to disk, creating parent directories as needed.
func Save(cfg *Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
