package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	reInterface  = regexp.MustCompile(`(?im)^\s*\[Interface\]`)
	rePeer       = regexp.MustCompile(`(?im)^\s*\[Peer\]`)
	rePrivateKey = regexp.MustCompile(`(?im)^\s*PrivateKey\s*=\s*\S+`)
	reAddress    = regexp.MustCompile(`(?im)^\s*Address\s*=\s*\S+`)
	rePublicKey  = regexp.MustCompile(`(?im)^\s*PublicKey\s*=\s*\S+`)
	reEndpoint   = regexp.MustCompile(`(?im)^\s*Endpoint\s*=\s*\S+`)
	reAllowedIPs = regexp.MustCompile(`(?im)^\s*AllowedIPs\s*=\s*\S+`)

	reSecret = regexp.MustCompile(`(?im)^(\s*(?:PrivateKey|PresharedKey|HeaderProtectionKey)\s*=\s*)(\S+)`)

	reAwg3   = regexp.MustCompile(`(?im)^\s*(?:HeaderProtectionKey|ContentPaddingAddition|RekeyAfterTime|RekeyTimeout|RejectAfterTime|KeepaliveTimeout|MaxHandshakeAttempts)\s*=`)
	reAwg2   = regexp.MustCompile(`(?im)^\s*I[1-5]\s*=`)
	reAwgOld = regexp.MustCompile(`(?im)^\s*(?:Jc|Jmin|Jmax|S[1-4]|H[1-4])\s*=`)
)

// validateConfig проверяет, что в конфиге есть обязательные поля WireGuard.
func validateConfig(config string) error {
	if strings.TrimSpace(config) == "" {
		return fmt.Errorf("файл конфигурации пуст")
	}
	checks := []struct {
		re   *regexp.Regexp
		name string
	}{
		{reInterface, "секция [Interface]"},
		{reAddress, "параметр Address"},
		{rePrivateKey, "параметр PrivateKey"},
		{rePeer, "секция [Peer]"},
		{rePublicKey, "параметр PublicKey"},
		{reEndpoint, "параметр Endpoint"},
		{reAllowedIPs, "параметр AllowedIPs"},
	}
	for _, c := range checks {
		if !c.re.MatchString(config) {
			return fmt.Errorf("в конфигурации отсутствует %s", c.name)
		}
	}
	return nil
}

// maskSecrets прячет значения приватных ключей для показа в интерфейсе.
func maskSecrets(config string) string {
	return reSecret.ReplaceAllString(config, "$1••••••••••••••••••••••••••••••••••••••••••••")
}

// detectProtocol определяет вариант протокола по набору параметров.
func detectProtocol(config string) string {
	switch {
	case reAwg3.MatchString(config):
		return "AmneziaWG 3.x"
	case reAwg2.MatchString(config):
		return "AmneziaWG 2.0"
	case reAwgOld.MatchString(config):
		return "AmneziaWG Legacy"
	default:
		return "WireGuard"
	}
}

// listProfiles возвращает имена сохранённых .conf-профилей.
func listProfiles() ([]string, error) {
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".conf") {
			names = append(names, e.Name())
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	return names, nil
}

// profilePath возвращает безопасный путь к профилю (без выхода из каталога).
func profilePath(name string) (string, error) {
	if name != filepath.Base(name) || name == "" {
		return "", fmt.Errorf("недопустимое имя профиля")
	}
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// importProfile validates raw config content read from an arbitrary source
// (a native file path or a URI reader) and stores it under the profiles
// directory, without overwriting an existing file of the same name.
// Returns the stored file name.
func importProfile(originalName string, raw []byte) (string, error) {
	config := normalizeConfig(string(raw))
	if err := validateConfig(config); err != nil {
		return "", err
	}

	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	if base == "" {
		base = "config"
	}
	name := base + ".conf"
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
			break
		}
		name = fmt.Sprintf("%s (%d).conf", base, i)
	}

	if err := writePrivateFile(filepath.Join(dir, name), []byte(config)); err != nil {
		return "", err
	}
	return name, nil
}

// writeProxyConfig generates the wireproxy engine config that references the
// chosen WireGuard profile and exposes it as a SOCKS5 proxy on the given port.
func writeProxyConfig(profilePath string, port int, listenOnLAN bool) (string, error) {
	root, err := appDir()
	if err != nil {
		return "", err
	}
	bindHost := "127.0.0.1"
	if listenOnLAN {
		bindHost = "0.0.0.0"
	}
	content := fmt.Sprintf("WGConfig = %s\n\n[Socks5]\nBindAddress = %s:%d\n", profilePath, bindHost, port)
	path := filepath.Join(root, "awgproxy.conf")
	if err := writePrivateFile(path, []byte(content)); err != nil {
		return "", err
	}
	return path, nil
}

func deleteProfile(name string) error {
	path, err := profilePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// normalizeConfig приводит переносы строк к \n, разворачивая экранированные "\n"
// из однострочных конфигов.
func normalizeConfig(config string) string {
	if !strings.ContainsAny(config, "\r\n") && strings.Contains(config, `\n`) {
		replacer := strings.NewReplacer(`\r\n`, "\n", `\n`, "\n", `\r`, "\n")
		config = replacer.Replace(config)
	}
	config = strings.ReplaceAll(config, "\r\n", "\n")
	config = strings.ReplaceAll(config, "\r", "\n")
	return config
}
