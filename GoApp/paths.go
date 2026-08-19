package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// appDir возвращает каталог данных приложения, создавая его при необходимости
// с правами только для владельца (0700).
func appDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = home
		}
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			base = xdg
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
	}

	dir := filepath.Join(base, "AtomicAWG")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	_ = os.Chmod(dir, 0o700)
	return dir, nil
}

func profilesDir() (string, error) {
	root, err := appDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "profiles")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	_ = os.Chmod(dir, 0o700)
	return dir, nil
}

func settingsPath() (string, error) {
	root, err := appDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "settings.json"), nil
}

// writePrivateFile пишет файл атомарно (tmp + rename) с правами 0600.
func writePrivateFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
