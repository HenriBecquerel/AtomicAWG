package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amnezia-vpn/amneziawg-go/v3/device"
	"github.com/awgproxy/awgproxy"
)

// tunnelStats mirrors the counters the UI displays.
type tunnelStats struct {
	ReceivedBytes int64
	SentBytes     int64
	LastHandshake time.Time
	HasHandshake  bool
}

// tunnelRuntime owns the embedded WireGuard/AmneziaWG engine and the proxy
// routines spawned from a config file. There is no external process: the
// core is a linked-in Go package, not a bundled second executable.
type tunnelRuntime struct {
	mu sync.Mutex
	vt *wireproxy.VirtualTun
}

func newTunnelRuntime() *tunnelRuntime {
	return &tunnelRuntime{}
}

func (t *tunnelRuntime) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.vt != nil
}

// Start parses the config at configPath and brings the tunnel plus every
// configured routine (SOCKS5, HTTP, SNI, ...) up. onLog receives every log
// line produced by the engine and by the proxy routines.
func (t *tunnelRuntime) Start(configPath string, onLog func(string)) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.vt != nil {
		return nil
	}

	conf, err := wireproxy.ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать конфигурацию ядра: %w", err)
	}

	logf := func(prefix string) func(string, ...any) {
		return func(format string, args ...any) {
			line := strings.TrimRight(fmt.Sprintf(format, args...), "\n")
			if line != "" && onLog != nil {
				onLog(prefix + line)
			}
		}
	}
	engineLogger := &device.Logger{
		Verbosef: logf(""),
		Errorf:   logf("ERROR: "),
	}
	wireproxy.SetLogOutput(lineWriter{onLog: onLog})

	vt, err := wireproxy.StartWireguardWithLogger(conf, engineLogger)
	if err != nil {
		return fmt.Errorf("не удалось поднять туннель: %w", err)
	}

	started := make(chan error, len(conf.Routines))
	for _, spawner := range conf.Routines {
		spawner := spawner
		go func() {
			err := spawner.SpawnRoutine(vt)
			// A nil error normally means "ran until Stop() closed the
			// listener"; report failures that happen before that so Start
			// can surface immediate bind errors (e.g. port already in use).
			select {
			case started <- err:
			default:
				if err != nil && onLog != nil {
					onLog("ERROR: " + err.Error())
				}
			}
		}()
	}

	// Give routines a brief moment to fail fast on bind errors before we
	// report success; a slow/blocking routine is treated as having started.
	select {
	case err := <-started:
		if err != nil {
			vt.Stop()
			return err
		}
	case <-time.After(300 * time.Millisecond):
	}

	vt.StartPingIPs()
	t.vt = vt
	return nil
}

func (t *tunnelRuntime) Stop() {
	t.mu.Lock()
	vt := t.vt
	t.vt = nil
	t.mu.Unlock()
	if vt != nil {
		vt.Stop()
	}
}

// Stats reads live counters straight from the WireGuard device's UAPI
// operation — no local network hop to a metrics endpoint is involved.
func (t *tunnelRuntime) Stats() (tunnelStats, error) {
	t.mu.Lock()
	vt := t.vt
	t.mu.Unlock()
	if vt == nil || vt.Dev == nil {
		return tunnelStats{}, fmt.Errorf("туннель не запущен")
	}

	raw, err := vt.Dev.IpcGet()
	if err != nil {
		return tunnelStats{}, err
	}

	var stats tunnelStats
	var lastHandshakeUnix int64
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "rx_bytes":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				stats.ReceivedBytes += n
			}
		case "tx_bytes":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				stats.SentBytes += n
			}
		case "last_handshake_time_sec":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil && n > lastHandshakeUnix {
				lastHandshakeUnix = n
			}
		}
	}
	if lastHandshakeUnix > 0 {
		stats.LastHandshake = time.Unix(lastHandshakeUnix, 0)
		stats.HasHandshake = true
	}
	return stats, nil
}

// lineWriter adapts the standard "log" package (used internally by the core
// for proxy-routine logs) to the UI's line callback.
type lineWriter struct {
	onLog func(string)
}

func (w lineWriter) Write(p []byte) (int, error) {
	if w.onLog != nil {
		if line := strings.TrimRight(string(p), "\n"); line != "" {
			w.onLog(line)
		}
	}
	return len(p), nil
}
