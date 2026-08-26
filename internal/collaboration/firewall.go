package collaboration

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
)

const (
	firewallRuleTCP = "SPH Manager TCP"
	firewallRuleUDP = "SPH Manager UDP"
)

// EnsureFirewallRules attempts to add Windows Firewall inbound rules for the
// collaboration ports. Returns a warning message if rules could not be created
// (typically because the app is not running as administrator).
func EnsureFirewallRules(tcpPort, udpPort int, log *slog.Logger) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	warnings := []string{}

	tryAdd := func(name, protocol string, port int) {
		cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
			"name="+name,
			"dir=in",
			"action=allow",
			"protocol="+fmt.Sprintf("%s", protocol),
			"localport="+fmt.Sprintf("%d", port),
			"enable=yes",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Warn("gagal membuat firewall rule",
				"rule", name, "error", err, "output", string(out))
			warnings = append(warnings, name)
		} else {
			log.Info("firewall rule siap", "rule", name, "port", port)
		}
	}

	tryAdd(firewallRuleTCP, "TCP", tcpPort)
	tryAdd(firewallRuleUDP, "UDP", udpPort)

	if len(warnings) > 0 {
		return "Firewall rules belum aktif. Jalankan aplikasi sebagai Administrator untuk " +
			"mengizinkan koneksi jaringan, atau buka port secara manual."
	}
	return ""
}

// RemoveFirewallRules removes the SPH Manager firewall rules.
func RemoveFirewallRules(log *slog.Logger) {
	if runtime.GOOS != "windows" {
		return
	}
	for _, name := range []string{firewallRuleTCP, firewallRuleUDP} {
		cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Warn("gagal hapus firewall rule", "rule", name, "error", err, "output", string(out))
		} else {
			log.Info("firewall rule dihapus", "rule", name)
		}
	}
}

// IsAdmin checks if the current process has administrator privileges on Windows.
func IsAdmin() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	cmd := exec.Command("net", "session")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "Session ID")
}
