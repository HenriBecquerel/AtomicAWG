package main

import (
	"encoding/json"
	"os"
)

type appSettings struct {
	ProxyPort           int    `json:"proxyPort"`
	ListenOnLAN         bool   `json:"listenOnLan"`
	ShowDebugLog        bool   `json:"showDebugLog"`
	SelectedProfileFile string `json:"selectedProfileFile"`
}

func defaultSettings() appSettings {
	return appSettings{ProxyPort: 1080}
}

func loadSettings() appSettings {
	path, err := settingsPath()
	if err != nil {
		return defaultSettings()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultSettings()
	}
	var s appSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return defaultSettings()
	}
	if s.ProxyPort < 1 || s.ProxyPort > 65535 {
		return defaultSettings()
	}
	return s
}

func (s appSettings) save() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writePrivateFile(path, data)
}
