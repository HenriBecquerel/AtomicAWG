package main

import (
	"net"
	"os"
	"testing"
	"time"
)

// TestSmokeEndToEnd is an integration check for the whole local pipeline:
// import a profile, generate the engine config, start the embedded engine
// (no subprocess involved) and confirm the SOCKS5 listener actually accepts
// TCP connections, then stop cleanly and confirm the port is released.
// It uses a temp HOME so it never touches the real app data directory.
func TestSmokeEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)

	sample := `[Interface]
Address = 10.30.0.2/32
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Jc = 5
Jmin = 20
Jmax = 50

[Peer]
PublicKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
AllowedIPs = 0.0.0.0/0
Endpoint = 127.0.0.1:51820
PersistentKeepalive = 25
`

	name, err := importProfile("smoke.conf", []byte(sample))
	if err != nil {
		t.Fatalf("importProfile: %v", err)
	}
	path, err := profilePath(name)
	if err != nil {
		t.Fatalf("profilePath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile not written: %v", err)
	}

	if !tcpPortAvailable(38080, false) {
		t.Fatal("expected port 38080 to be free before start")
	}

	proxyConfigPath, err := writeProxyConfig(path, 38080, false)
	if err != nil {
		t.Fatalf("writeProxyConfig: %v", err)
	}

	rt := newTunnelRuntime()
	var logLines []string
	if err := rt.Start(proxyConfigPath, func(line string) { logLines = append(logLines, line) }); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer rt.Stop()

	if !rt.IsRunning() {
		t.Fatal("expected runtime to report running")
	}

	conn, err := net.DialTimeout("tcp", "127.0.0.1:38080", 2*time.Second)
	if err != nil {
		t.Fatalf("SOCKS5 listener did not accept a connection: %v", err)
	}
	_ = conn.Close()

	if _, err := rt.Stats(); err != nil {
		t.Fatalf("Stats: %v", err)
	}

	rt.Stop()
	if rt.IsRunning() {
		t.Fatal("expected runtime to report stopped after Stop")
	}

	// Port must be free again after Stop.
	deadline := time.Now().Add(2 * time.Second)
	for !tcpPortAvailable(38080, false) {
		if time.Now().After(deadline) {
			t.Fatal("port 38080 still bound after Stop")
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Logf("captured %d log lines, e.g. %v", len(logLines), logLines)
}
