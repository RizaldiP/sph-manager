package collaboration

import (
	"encoding/json"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// announcePacket: muatan UDP yang di-broadcast host saat room aktif.
type announcePacket struct {
	RoomID         string `json:"roomId"`
	RoomName       string `json:"roomName"`
	DocumentNumber string `json:"documentNumber"`
	ProjectName    string `json:"projectName"`
	HostName       string `json:"hostName"`
	WSPort         int    `json:"port"`
	Users          int    `json:"users"`
}

// ===== Announcer (sisi host) =====

// Announcer menyiarkan paket room secara periodik ke alamat broadcast interface.
type Announcer struct {
	conn     *net.UDPConn
	dests    []*net.UDPAddr
	interval time.Duration
	packet   atomic.Value // announcePacket
	log      *slog.Logger
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
}

// startAnnouncer menyiapkan socket broadcast; pengiriman berjalan pada goroutine.
func startAnnouncer(port int, interval time.Duration, log *slog.Logger) (*Announcer, error) {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: port})
	if err != nil {
		return nil, err
	}

	// Set SO_BROADCAST agar Windows mengizinkan pengiriman ke alamat broadcast.
	if sc, err := conn.SyscallConn(); err == nil {
		_ = sc.Control(func(fd uintptr) {
			setBroadcast(fd)
		})
	}

	dests := addIfaceBroadcasts(port)

	a := &Announcer{
		conn:     conn,
		dests:    dests,
		interval: interval,
		log:      log,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	go a.loop()
	return a, nil
}

// addIfaceBroadcasts mengembalikan daftar alamat broadcast: global 255.255.255.255
// ditambah directed broadcast tiap interface IPv4 aktif.
func addIfaceBroadcasts(port int) []*net.UDPAddr {
	dests := []*net.UDPAddr{{IP: net.IPv4(255, 255, 255, 255), Port: port}}
	ifaces, err := net.Interfaces()
	if err != nil {
		return dests
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
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
			broadcast := broadcastOf(ipnet)
			if broadcast == nil {
				continue
			}
			dup := false
			for _, d := range dests {
				if d.IP.Equal(broadcast) {
					dup = true
					break
				}
			}
			if !dup {
				dests = append(dests, &net.UDPAddr{IP: broadcast, Port: port})
			}
		}
	}
	return dests
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

func (a *Announcer) loop() {
	defer close(a.doneCh)
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
		for _, d := range a.dests {
			if _, wErr := a.conn.WriteToUDP(b, d); wErr != nil {
				a.log.Warn("gagal broadcast discovery", "dest", d.IP.String(), "error", wErr)
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
		_ = a.conn.Close()
		<-a.doneCh
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
		if json.Unmarshal(buf[:n], &p) != nil || p.RoomID == "" || p.WSPort <= 0 {
			continue
		}
		hostIP := ""
		if addr != nil {
			hostIP = addr.IP.String()
		}
		l.mu.Lock()
		l.rooms[p.RoomID] = DiscoveredRoom{
			RoomID:         p.RoomID,
			RoomName:       p.RoomName,
			DocumentNumber: p.DocumentNumber,
			ProjectName:    p.ProjectName,
			HostIP:         hostIP,
			HostName:       p.HostName,
			Port:           p.WSPort,
			Users:          p.Users,
			LastSeen:       time.Now(),
		}
		l.mu.Unlock()
	}
}

func (l *Listener) pruneLoop(interval time.Duration) {
	defer close(l.pruneDone)
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
