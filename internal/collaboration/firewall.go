package collaboration

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
)

const (
	firewallRuleTCP = "SPH Manager TCP"
	firewallRuleUDP = "SPH Manager UDP"
)

// EnsureFirewallRules attempts to add Windows Firewall inbound rules for the
// collaboration ports. It uses netsh which may require admin privileges.
// Returns nil if rules already exist or were added successfully.
func EnsureFirewallRules(tcpPort, udpPort int, log *slog.Logger) error {
	if runtime.GOOS != "windows" {
		return nil
	}

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
		} else {
			log.Info("firewall rule siap", "rule", name, "port", port)
		}
	}

	tryAdd(firewallRuleTCP, "TCP", tcpPort)
	tryAdd(firewallRuleUDP, "UDP", udpPort)

	return nil
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
