package collaboration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// announcePacket: muatan UDP yang di-broadcast host saat room aktif.
// HostIPs berisi semua alamat IP host yang reachable, sehingga client bisa
// mencoba IP selain source address paket (penting saat host punya banyak
// interface aktif: kabel LAN + WiFi sekaligus).
type announcePacket struct {
	RoomID         string   `json:"roomId"`
	RoomName       string   `json:"roomName"`
	DocumentNumber string   `json:"documentNumber"`
	ProjectName    string   `json:"projectName"`
	HostName       string   `json:"hostName"`
	WSPort         int      `json:"port"`
	Users          int      `json:"users"`
	HostIPs        []string `json:"hostIPs,omitempty"`
	Status         string   `json:"status,omitempty"`
}

// ===== Announcer (sisi host) =====

// announceTarget: satu tujuan broadcast beserta source IP socket pengirim.
// srcIP kosong berarti memakai socket wildcard (untuk broadcast global).
type announceTarget struct {
	srcIP string
	dst   *net.UDPAddr
}

// announceConn: satu socket UDP yang mengirim ke satu alamat broadcast.
type announceConn struct {
	conn  *net.UDPConn
	dst   *net.UDPAddr
	srcIP string
}

// Announcer menyiarkan paket room secara periodik ke alamat broadcast interface.
// Setiap interface IPv4 fisik/aktif mendapat socket sendiri yang dibind ke IP
// interface tersebut, sehingga (a) tiap subnet pasti menerima paket dari
// interface-nya dan (b) source address paket selalu IP yang benar-benar
// reachable — bukan IP interface default-route yang kebetulan dipilih OS.
type Announcer struct {
	anns     []*announceConn
	interval time.Duration
	packet   atomic.Value // announcePacket
	log      *slog.Logger
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
}

// startAnnouncer menyiapkan socket-socket broadcast; pengiriman di goroutine.
func startAnnouncer(port int, interval time.Duration, log *slog.Logger) (*Announcer, error) {
	anns, err := newAnnounceConns(port, log)
	if err != nil {
		return nil, err
	}
	if len(anns) == 0 {
		log.Warn("tidak ada interface IPv4 aktif untuk broadcast discovery (join manual tetap bisa)")
	}
	a := &Announcer{
		anns:     anns,
		interval: interval,
		log:      log,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	go a.loop()
	return a, nil
}

// newAnnounceConns membuka satu socket UDP per target broadcast, masing-masing
// dibind ke IP interface terkait agar source address terjamin benar.
func newAnnounceConns(port int, log *slog.Logger) ([]*announceConn, error) {
	targets := buildAnnounceTargets(port)
	var out []*announceConn
	for _, t := range targets {
		var bindsrc net.IP
		if t.srcIP != "" {
			bindsrc = net.ParseIP(t.srcIP)
		}
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: bindsrc, Port: 0})
		if err != nil {
			log.Warn("gagal membuka socket broadcast", "src", t.srcIP, "error", err)
			continue
		}
		// Set SO_BROADCAST agar Windows mengizinkan pengiriman ke alamat broadcast.
		if sc, err := conn.SyscallConn(); err == nil {
			_ = sc.Control(func(fd uintptr) {
				setBroadcast(fd)
			})
		}
		out = append(out, &announceConn{conn: conn, dst: t.dst, srcIP: t.srcIP})
	}
	if len(out) == 0 && len(targets) > 0 {
		return nil, fmt.Errorf("gagal membuka socket UDP broadcast")
	}
	return out, nil
}

// buildAnnounceTargets mengembalikan daftar target broadcast: global 255.255.255.255
// ditambah directed broadcast tiap interface IPv4 non-virtual aktif.
func buildAnnounceTargets(port int) []announceTarget {
	targets := []announceTarget{{srcIP: "", dst: &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: port}}}
	seen := map[string]bool{}
	for _, ii := range listIPv4Interfaces() {
		broadcast := broadcastOf(ii.ipnet)
		if broadcast == nil {
			continue
		}
		src := ii.ip.String()
		key := src + "|" + broadcast.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		targets = append(targets, announceTarget{srcIP: src, dst: &net.UDPAddr{IP: broadcast, Port: port}})
	}
	return targets
}

func broadcastOf(ipnet *net.IPNet) net.IP {
	ip4 := ipnet.IP.To4()
	mask4 := ipnet.Mask
	if len(mask4) != 4 {
		return nil
	}
	out := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		out[i] = ip4[i] | ^mask4[i]
	}
	return out
}

func (a *Announcer) Set(p announcePacket) { a.packet.Store(p) }

func (a *Announcer) sendPacket(p announcePacket) {
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	for _, ac := range a.anns {
		_, _ = ac.conn.WriteToUDP(b, ac.dst)
	}
}

func (a *Announcer) loop() {
	defer close(a.doneCh)
	defer func() {
		if rec := recover(); rec != nil {
			a.log.Error("announcer.loop panic", "recover", rec)
		}
	}()
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	sendOnce := func() {
		p, _ := a.packet.Load().(announcePacket)
		if p.RoomID == "" {
			return
		}
		b, err := json.Marshal(p)
		if err != nil {
			a.log.Warn("gagal marshal announce packet", "error", err)
			return
		}
		for _, ac := range a.anns {
			if _, wErr := ac.conn.WriteToUDP(b, ac.dst); wErr != nil {
				a.log.Warn("gagal broadcast discovery", "dest", ac.dst.IP.String(), "src", ac.srcIP, "error", wErr)
			}
		}
	}
	sendOnce()
	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			sendOnce()
		}
	}
}

func (a *Announcer) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
		<-a.doneCh
		if last, ok := a.packet.Load().(announcePacket); ok && last.RoomID != "" {
			goodbye := announcePacket{RoomID: last.RoomID, Status: "CLOSED"}
			for i := 0; i < 3; i++ {
				a.sendPacket(goodbye)
				if i < 2 {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
		for _, ac := range a.anns {
			_ = ac.conn.Close()
		}
	})
}

// ===== Listener (sisi client/lobby) =====

// Listener mendengarkan broadcast discovery dan memelihara daftar room hidup.
type Listener struct {
	conn     *net.UDPConn
	log      *slog.Logger
	mu       sync.Mutex
	rooms    map[string]DiscoveredRoom
	ttl      time.Duration
	onDead   func() // dipanggil saat readLoop berhenti karena error
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
	pruneDone chan struct{}
}

func startListener(port int, ttl time.Duration, pruneInterval time.Duration, log *slog.Logger) (*Listener, error) {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: port}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}
	l := &Listener{
		conn:      conn,
		log:       log,
		rooms:     map[string]DiscoveredRoom{},
		ttl:       ttl,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
		pruneDone: make(chan struct{}),
	}
	go l.readLoop()
	go l.pruneLoop(pruneInterval)
	return l, nil
}

func (l *Listener) readLoop() {
	defer close(l.doneCh)
	defer func() {
		if rec := recover(); rec != nil {
			l.log.Error("listener.readLoop panic", "recover", rec)
		}
	}()
	deadCalled := false
	defer func() {
		if !deadCalled && l.onDead != nil {
			go l.onDead()
		}
	}()
	buf := make([]byte, 4096)
	for {
		n, addr, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-l.stopCh:
				deadCalled = true
				return
			default:
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				l.log.Warn("discovery listener readLoop berhenti", "error", err)
				return
			}
		}
		var p announcePacket
		if err := json.Unmarshal(buf[:n], &p); err != nil || p.RoomID == "" {
			continue
		}
		// Goodbye packet: room ditutup, langsung hapus dari daftar.
		if p.Status == "CLOSED" {
			l.mu.Lock()
			delete(l.rooms, p.RoomID)
			l.mu.Unlock()
			continue
		}
		if p.WSPort <= 0 {
			continue
		}
		hostIPs, primary := candidatesFromAnnounce(p, addr)
		l.mu.Lock()
		l.rooms[p.RoomID] = DiscoveredRoom{
			RoomID:         p.RoomID,
			RoomName:       p.RoomName,
			DocumentNumber: p.DocumentNumber,
			ProjectName:    p.ProjectName,
			HostIP:         primary,
			HostIPs:        hostIPs,
			HostName:       p.HostName,
			Port:           p.WSPort,
			Users:          p.Users,
			LastSeen:       time.Now(),
		}
		l.mu.Unlock()
	}
}

// candidatesFromAnnounce mengembalikan daftar IP host yang akan dicoba client:
// source address paket (bila valid) lebih dulu, lalu daftar HostIPs dari payload.
// Menoleransi source 0.0.0.0 dan membuang IP loopback/link-local/tidak valid.
func candidatesFromAnnounce(p announcePacket, addr *net.UDPAddr) ([]string, string) {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		ip := net.ParseIP(s)
		if ip == nil || ip.To4() == nil {
			return
		}
		ip = ip.To4()
		if ip.IsLoopback() || ip.IsUnspecified() {
			return
		}
		if ip[0] == 169 && ip[1] == 254 {
			return
		}
		key := ip.String()
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, key)
	}
	src := ""
	if addr != nil && addr.IP != nil {
		src = addr.IP.String()
	}
	add(src)
	for _, s := range p.HostIPs {
		add(s)
	}
	primary := ""
	if len(out) > 0 {
		primary = out[0]
	}
	return out, primary
}

func (l *Listener) pruneLoop(interval time.Duration) {
	defer close(l.pruneDone)
	defer func() {
		if rec := recover(); rec != nil {
			l.log.Error("listener.pruneLoop panic", "recover", rec)
		}
	}()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-l.ttl)
			l.mu.Lock()
			for id, r := range l.rooms {
				if r.LastSeen.Before(cutoff) {
					delete(l.rooms, id)
				}
			}
			l.mu.Unlock()
		}
	}
}

// Rooms mengembalikan snapshot daftar room terurut dari yang terbaru terlihat.
func (l *Listener) Rooms() []DiscoveredRoom {
	l.mu.Lock()
	out := make([]DiscoveredRoom, 0, len(l.rooms))
	for _, r := range l.rooms {
		out = append(out, r)
	}
	l.mu.Unlock()
	sortDiscoveredByNewest(out)
	return out
}

func (l *Listener) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopCh)
		_ = l.conn.Close()
		<-l.doneCh
		<-l.pruneDone
	})
}
