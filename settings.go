package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// Settings represents persisted application settings
type Settings struct {
	MIDIPort        int              `json:"midiPort"`
	ChannelMappings map[string]uint8 `json:"channelMappings"`
}

// getSettingsPath returns the path to the settings file
func getSettingsPath() (string, error) {
	return xdg.ConfigFile("forbidden_sequencer/settings.json")
}

// LoadSettings loads settings from disk, returns defaults if file doesn't exist
func LoadSettings() (*Settings, error) {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return nil, err
	}

	// Return defaults if file doesn't exist
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return &Settings{
			MIDIPort:        0,
			ChannelMappings: make(map[string]uint8),
		}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	if settings.ChannelMappings == nil {
		settings.ChannelMappings = make(map[string]uint8)
	}

	return &settings, nil
}

// SaveSettings saves settings to disk
func SaveSettings(settings *Settings) error {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}
