package collaboration

import (
	"net"
	"testing"
)

func TestCandidatesFromAnnounce(t *testing.T) {
	tests := []struct {
		name      string
		packet    announcePacket
		addr      *net.UDPAddr
		wantHosts []string
		wantPrime string
	}{
		{
			name:      "source address diikuti HostIPs payload, tanpa duplikat",
			packet:    announcePacket{HostIPs: []string{"192.168.1.15", "192.168.1.10", "192.168.1.15"}},
			addr:      &net.UDPAddr{IP: net.ParseIP("192.168.1.10")},
			wantHosts: []string{"192.168.1.10", "192.168.1.15"},
			wantPrime: "192.168.1.10",
		},
		{
			name:      "source 0.0.0.0 diabaikan, pakai daftar payload",
			packet:    announcePacket{HostIPs: []string{"10.0.0.5"}},
			addr:      &net.UDPAddr{IP: net.ParseIP("0.0.0.0")},
			wantHosts: []string{"10.0.0.5"},
			wantPrime: "10.0.0.5",
		},
		{
			name:      "loopback, link-local, dan IP tidak valid dibuang",
			packet:    announcePacket{HostIPs: []string{"127.0.0.1", "169.254.1.2", "not-an-ip", "192.168.0.7"}},
			addr:      &net.UDPAddr{IP: net.ParseIP("169.254.10.10")},
			wantHosts: []string{"192.168.0.7"},
			wantPrime: "192.168.0.7",
		},
		{
			name:      "tidak ada sumber sama sekali",
			packet:    announcePacket{},
			addr:      nil,
			wantHosts: nil,
			wantPrime: "",
		},
		{
			name:      "IPv6 tidak dipakai (hanya IPv4)",
			packet:    announcePacket{HostIPs: []string{"fe80::1", "192.168.1.9"}},
			addr:      &net.UDPAddr{IP: net.ParseIP("fe80::2")},
			wantHosts: []string{"192.168.1.9"},
			wantPrime: "192.168.1.9",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hosts, prime := candidatesFromAnnounce(tc.packet, tc.addr)
			if prime != tc.wantPrime {
				t.Errorf("primary = %q, mau %q (hosts=%v)", prime, tc.wantPrime, hosts)
			}
			if len(hosts) != len(tc.wantHosts) {
				t.Fatalf("hosts = %v, mau %v", hosts, tc.wantHosts)
			}
			for i := range hosts {
				if hosts[i] != tc.wantHosts[i] {
					t.Errorf("hosts[%d] = %q, mau %q (semua=%v)", i, hosts[i], tc.wantHosts[i], hosts)
				}
			}
		})
	}
}

func TestBroadcastOf(t *testing.T) {
	tests := []struct {
		ip   string
		mask string
		want string
	}{
		{"192.168.1.10", "255.255.255.0", "192.168.1.255"},
		{"10.0.0.5", "255.0.0.0", "10.255.255.255"},
		{"172.16.4.9", "255.255.0.0", "172.16.255.255"},
	}
	for _, tc := range tests {
		ipnet := &net.IPNet{IP: net.ParseIP(tc.ip), Mask: net.IPMask(net.ParseIP(tc.mask).To4())}
		got := broadcastOf(ipnet)
		if got == nil || got.String() != tc.want {
			t.Errorf("broadcastOf(%s/%s) = %v, mau %s", tc.ip, tc.mask, got, tc.want)
		}
	}
}

func TestBuildAnnounceTargetsAlwaysHasGlobal(t *testing.T) {
	targets := buildAnnounceTargets(DefaultDiscoveryPort)
	if len(targets) == 0 {
		t.Fatal("harus selalu ada minimal target global broadcast")
	}
	if targets[0].srcIP != "" || targets[0].dst.Port != DefaultDiscoveryPort {
		t.Errorf("target pertama harus global broadcast: %+v", targets[0])
	}
}