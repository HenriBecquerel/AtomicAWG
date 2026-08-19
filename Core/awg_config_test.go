package wireproxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amnezia-vpn/amneziawg-go/v3/conn"
	"github.com/amnezia-vpn/amneziawg-go/v3/device"
	"github.com/amnezia-vpn/amneziawg-go/v3/tun/netstack"
)

func TestAWG3ConfigBecomesUAPI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "awg.conf")
	config := `[Interface]
Address = 10.10.0.2/32
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Jc = 5
Jmin = 20
Jmax = 50
S1 = 12
S2 = 12
S3 = 12
S4 = 12
H1 = 100-200
H2 = 300-400
H3 = 500-600
H4 = 700-800
I1 = <r 16>
HeaderProtectionKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
ContentPaddingAddition = 8-32
RekeyAfterTime = 100-120

[Peer]
PublicKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
AllowedIPs = 0.0.0.0/0
Endpoint = 127.0.0.1:51820
PersistentKeepalive = 20-30

[Socks5]
BindAddress = 127.0.0.1:1080
`
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	setting, err := CreateIPCRequest(parsed.Device)
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"jc=5", "s4=12", "h1=100-200", "i1=<r 16>",
		"header_protection_key=0000000000000000000000000000000000000000000000000000000000000000",
		"content_padding_addition=8-32", "rekey_after_time=100-120",
		"persistent_keepalive_interval=20-30",
	} {
		if !strings.Contains(setting.IpcRequest, expected+"\n") {
			t.Fatalf("IPC request does not contain %q:\n%s", expected, setting.IpcRequest)
		}
	}

	tun, _, err := netstack.CreateNetTUN(setting.DeviceAddr, setting.DNS, setting.MTU)
	if err != nil {
		t.Fatal(err)
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelSilent, ""))
	defer dev.Close()
	if err := dev.IpcSet(setting.IpcRequest); err != nil {
		t.Fatalf("AmneziaWG engine rejected generated UAPI configuration: %v", err)
	}
}

func TestPlainWireGuardHasNoAWGParameters(t *testing.T) {
	conf := &DeviceConfig{SecretKey: strings.Repeat("0", 64), AWGParameters: map[string]string{}}
	setting, err := CreateIPCRequest(conf)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"jc=", "h1=", "i1=", "header_protection_key="} {
		if strings.Contains(setting.IpcRequest, forbidden) {
			t.Fatalf("plain WireGuard request unexpectedly contains %q", forbidden)
		}
	}
}
