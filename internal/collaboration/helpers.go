package collaboration

import (
	"crypto/rand"
	"crypto/subtle"
	"math/big"
	"net"
	"strings"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// equalConstTime membandingkan dua string tanpa timing leak (untuk access code).
func equalConstTime(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

const codeAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// GenerateRoomCode membuat kode room 6 karakter yang mudah dibaca/dikomunikasikan.
func GenerateRoomCode() string { return randomFrom(codeAlphabet, 6) }

// GenerateAccessCode membuat access code numerik 6 digit (§10.25).
func GenerateAccessCode() string { return randomFrom("0123456789", 6) }

func randomFrom(alphabet string, n int) string {
	out := make([]byte, n)
	max := big.NewInt(int64(len(alphabet)))
	for i := range out {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = alphabet[i%len(alphabet)]
			continue
		}
		out[i] = alphabet[v.Int64()]
	}
	return string(out)
}

// sanitizeIdentity membersihkan nama tampilan/device dari input pengguna.
func sanitizeIdentity(s string) string {
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	if len(s) > 100 {
		r := []rune(s)
		if len(r) > 100 {
			s = string(r[:100])
		}
	}
	return s
}

// statusLabelID menerjemahkan kode status dokumen ke label Indonesia untuk pesan error.
func statusLabelID(status string) string {
	switch status {
	case models.StatusDraft:
		return "Draft"
	case models.StatusReview:
		return "Review"
	case models.StatusFinal:
		return "Final"
	case models.StatusSent:
		return "Terkirim"
	case models.StatusAccepted:
		return "Disetujui"
	case models.StatusRejected:
		return "Ditolak"
	case models.StatusCancelled:
		return "Dibatalkan"
	default:
		return status
	}
}

// ifaceIPv4 menggambarkan satu alamat IPv4 pada interface jaringan aktif.
type ifaceIPv4 struct {
	ip    net.IP
	ipnet *net.IPNet
}

// listIPv4Interfaces mengembalikan alamat IPv4 semua interface non-virtual yang
// aktif (bukan loopback, bukan APIPA/link-local). dipakai untuk broadcast
// discovery, daftar IP host (toolbar), dan penapisan adapter virtual
// (Docker/WSL/VMware/VirtualBox/Hyper-V/dll.) yang tidak dapat dijangkau client.
func listIPv4Interfaces() []ifaceIPv4 {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []ifaceIPv4
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if isVirtualIface(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ip := ipnet.IP.To4()
			if ip.IsUnspecified() {
				continue
			}
			if ip[0] == 127 || (ip[0] == 169 && ip[1] == 254) {
				continue
			}
			out = append(out, ifaceIPv4{ip: append(net.IP(nil), ip...), ipnet: ipnet})
		}
	}
	return out
}

// isVirtualIface menandai adapter jaringan virtual (WSL, Hyper-V, Docker,
// VMware, VirtualBox, NPcap loopback, VPN, WiFi-Direct/ICS) agar tidak dipakai
// untuk broadcast maupun ditampilkan sebagai IP host yang dapat dijangkau.
func isVirtualIface(iface net.Interface) bool {
	name := strings.ToLower(iface.Name)
	for _, m := range []string{
		"vethernet", "wsl", "docker", "vmnet", "vmware", "virtualbox", "vbox",
		"tailscale", "zerotier", "hamachi", "npcap", "loopback", "kvm",
		"default switch", "microsoft wi-fi direct", "tap-windows", "p-tap",
		"utun", "tun", "tap", "bluetooth",
	} {
		if strings.Contains(name, m) {
			return true
		}
	}
	// Windows menamai adapter virtual ICS/WiFi-Direct memakai akhiran "*" atau "#".
	if strings.Contains(name, "*") || strings.Contains(name, "#") {
		return true
	}
	hw := strings.ToLower(iface.HardwareAddr.String())
	hw = strings.ReplaceAll(hw, ":", "")
	if hw == "" {
		return true
	}
	for _, o := range []string{
		"00155d", // WSL / Hyper-V
		"0003ff", // Hyper-V
		"005056", // VMware ESX
		"000569", // VMware
		"000c29", // VMware
		"080027", // VirtualBox
		"0a0027", // VirtualBox
		"0242",   // Docker
		"00163e", // Xen
		"525400", // QEMU
		"001c42", // Parallels
	} {
		if strings.HasPrefix(hw, o) {
			return true
		}
	}
	return false
}

// localIPs mengembalikan daftar IPv4 address aktif yang benar-benar dapat
// dijangkau (interface fisik saja; loopback 127.x, link-local 169.254.x, dan
// adapter virtual dibuang).
func localIPs() []string {
	var ips []string
	seen := map[string]bool{}
	for _, ii := range listIPv4Interfaces() {
		ip4 := ii.ip.String()
		if !seen[ip4] {
			seen[ip4] = true
			ips = append(ips, ip4)
		}
	}
	return ips
}

func cloneParticipants(src []Participant) []Participant {
	if src == nil {
		return nil
	}
	out := make([]Participant, len(src))
	copy(out, src)
	return out
}

func cloneActivities(src []services.CollabActivity) []services.CollabActivity {
	if src == nil {
		return nil
	}
	out := make([]services.CollabActivity, len(src))
	copy(out, src)
	return out
}

func cloneRoomInfo(src *RoomInfo) *RoomInfo {
	if src == nil {
		return nil
	}
	c := *src
	c.Participants = cloneParticipants(src.Participants)
	return &c
}

func cloneChat(src []ChatPayload) []ChatPayload {
	if src == nil {
		return nil
	}
	out := make([]ChatPayload, len(src))
	copy(out, src)
	return out
}

// sortDiscoveredByNewest mengurutkan entri discovery dari yang terakhir terlihat.
func sortDiscoveredByNewest(rows []DiscoveredRoom) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].LastSeen.After(rows[j-1].LastSeen); j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
