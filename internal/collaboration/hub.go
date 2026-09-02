package collaboration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// Config mengatur parameter jaringan/kolaborasi; nilai nol memakai default.
type Config struct {
	DeviceName       string
	DisableDiscovery bool          // matikan UDP announce/listen (untuk test)
	Heartbeat        time.Duration // interval PING client → host
	ReadWait         time.Duration // read deadline per pesan
	JoinWait         time.Duration // batas menunggu JOIN_REQUEST pertama
	PresenceTTL      time.Duration // batas diam peserta dianggap offline
	BackoffMin       time.Duration // awal reconnect backoff
	BackoffMax       time.Duration // plafon reconnect backoff
	DialTimeout      time.Duration // timeout koneksi pertama saat join
	AnnounceInterval time.Duration // interval UDP broadcast room
	DiscoverTTL      time.Duration // masa hidup entri discovered room
	ActivityCap      int
}

func (c Config) withDefaults() Config {
	if c.DeviceName == "" {
		if h, err := os.Hostname(); err == nil && h != "" {
			c.DeviceName = h
		} else {
			c.DeviceName = "PC"
		}
	}
	set := func(d *time.Duration, def time.Duration) {
		if *d <= 0 {
			*d = def
		}
	}
	set(&c.Heartbeat, 5*time.Second)
	set(&c.ReadWait, 20*time.Second)
	set(&c.JoinWait, 15*time.Second)
	set(&c.PresenceTTL, 30*time.Second)
	set(&c.BackoffMin, 500*time.Millisecond)
	set(&c.BackoffMax, 8*time.Second)
	set(&c.DialTimeout, 5*time.Second)
	set(&c.AnnounceInterval, 2*time.Second)
	set(&c.DiscoverTTL, 7*time.Second)
	if c.ActivityCap <= 0 {
		c.ActivityCap = 100
	}
	return c
}

// UISnapshot adalah potret penuh sesi kolaborasi yang dikirim ke frontend lewat
// event Wails `collab:sync`. Frontend cukup mengganti seluruh state-nya.
type UISnapshot struct {
	Mode         string                    `json:"mode"`
	Connection   string                    `json:"connection,omitempty"`
	Room         *RoomInfo                 `json:"room,omitempty"`
	Doc          json.RawMessage           `json:"doc,omitempty"`
	Participants []Participant             `json:"participants,omitempty"`
	Activities   []services.CollabActivity `json:"activities,omitempty"`
	Turn         *TurnState                `json:"turn,omitempty"`
	Chat         []ChatPayload             `json:"chat,omitempty"`
	Unread       int                       `json:"unread,omitempty"`
	Version      uint64                    `json:"version,omitempty"`
	Error        string                    `json:"error,omitempty"`
	Notice       string                    `json:"notice,omitempty"`
	Incoming     int                       `json:"incoming,omitempty"`
	MasterStatus []MasterStatusEntry       `json:"masterStatus,omitempty"`
}

// Emit dipanggil setiap ada perubahan sesi; wiring aplikasi meneruskannya ke UI.
type Emit func(UISnapshot)

// Manager mengatur satu sesi kolaborasi aktif per aplikasi: menjadi host sebuah
// room ATAU menjadi client pada room lain — tidak keduanya sekaligus.
// Manager mengimplementasikan services.RoomGuard sehingga dokumen yang sedang
// dibuka dalam room terkunci dari jalur solo.
type Manager struct {
	cfg  Config
	log  *slog.Logger
	ops  *services.CollabOps
	sph  *services.SphService
	chat ChatStore
	md   MasterDataStore

	mu       sync.Mutex
	emitFn   Emit
	room     *Room
	client   *Client
	listener *Listener

	// cache sisi client
	connStatus     string
	clientErr      string
	clientNotice   string
	clientVersion  uint64
	clientRoomMeta *RoomInfo
	clientDoc      json.RawMessage
	clientParts    []Participant
	clientActs     []services.CollabActivity
	clientTurn     *TurnState
	clientChat     []ChatPayload
	clientUnread   int
	hostUnread     int
	clientIncoming []MasterDataTransferPayload // master data masuk (sisi client) menunggu dipasang
	clientMasterStatus map[string][]MasterStatusEntry // packageID → status per penerima (sisi client)
}

func NewManager(cfg Config, ops *services.CollabOps, sph *services.SphService, log *slog.Logger) *Manager {
	return &Manager{
		cfg: cfg.withDefaults(),
		log: log,
		ops: ops,
		sph: sph,
	}
}

// SetChatStore memasang penyimpanan riwayat chat (host). Dipanggil app saat startup.
func (m *Manager) SetChatStore(cs ChatStore) {
	m.mu.Lock()
	m.chat = cs
	m.mu.Unlock()
}

// SetMasterDataStore memasang penyimpanan Master Data (inbox/sent). Dipanggil app saat startup.
func (m *Manager) SetMasterDataStore(st MasterDataStore) {
	m.mu.Lock()
	m.md = st
	m.mu.Unlock()
}

// SetEmit memasang kanal notifikasi UI. Dipanggil setelah startup Wails.
func (m *Manager) SetEmit(fn Emit) {
	m.mu.Lock()
	m.emitFn = fn
	m.mu.Unlock()
}

func (m *Manager) emit(snap UISnapshot) {
	m.mu.Lock()
	fn := m.emitFn
	m.mu.Unlock()
	if fn != nil {
		fn(snap)
	}
}

// ===== Guard (dipakai SphService) =====

// IsDocLocked melaporkan apakah dokumen sedang dibuka dalam room aktif.
func (m *Manager) IsDocLocked(sphDocumentID uint) bool {
	m.mu.Lock()
	r := m.room
	m.mu.Unlock()
	return r != nil && r.docID == sphDocumentID && !r.isClosed()
}

// ===== HOST =====

// HostRoom membuat room baru untuk draft SPH beserta WebSocket server-nya.
func (m *Manager) HostRoom(docID uint, roomName, displayName string, port int) (*RoomInfo, error) {
	displayName = sanitizeIdentity(displayName)
	if displayName == "" {
		displayName = m.cfg.DeviceName
	}

	doc, err := m.sph.Get(docID)
	if err != nil {
		return nil, err
	}
	if doc.Status != models.StatusDraft {
		return nil, services.NewConflictError(
			"Room hanya dapat dibuat untuk dokumen berstatus Draft (status dokumen ini: %s).",
			statusLabelID(doc.Status))
	}

	m.mu.Lock()
	if m.room != nil {
		m.mu.Unlock()
		return nil, services.NewConflictError("Aplikasi ini sudah menjadi host room \"%s\". Tutup room tersebut lebih dulu.", m.room.info.RoomName)
	}
	if m.client != nil {
		m.mu.Unlock()
		return nil, services.NewConflictError("Aplikasi ini sedang join ke room lain. Keluar dari room tersebut lebih dulu.")
	}

	srv, err := startWSServer(port, m.log, m.handleIncomingConn)
	if err != nil {
		m.mu.Unlock()
		m.log.Error("gagal membuka port kolaborasi", "port", port, "error", err)
		return nil, fmt.Errorf("gagal membuka port %d untuk room kolaborasi. "+
			"Izinkan \"SPH Manager\" pada jaringan privat bila Windows menanyakan, "+
			"atau gunakan port lain di halaman Pengaturan.", port)
	}

	firewallWarn := EnsureFirewallRules(srv.Port(), DefaultDiscoveryPort, m.log)

	now := time.Now()
	hostPart := Participant{
		ID:          uuid.NewString(),
		DisplayName: displayName,
		DeviceName:  m.cfg.DeviceName,
		Role:        RoleHost,
		JoinedAt:    now,
		LastSeen:    now,
	}
	info := RoomInfo{
		RoomID:          uuid.NewString(),
		SphDocumentID:   doc.ID,
		DocumentNumber:  doc.DocumentNumber,
		ProjectName:     doc.ProjectName,
		RoomCode:        GenerateRoomCode(),
		RoomName:        strings.TrimSpace(roomName),
		AccessCode:      GenerateAccessCode(),
		HostName:        displayName,
		HostDevice:      m.cfg.DeviceName,
		HostIPs:         localIPs(),
		Port:            srv.Port(),
		Status:          RoomStatusActive,
		FirewallWarning: firewallWarn,
		CreatedAt:       now,
		Participants:    []Participant{hostPart},
	}
	if info.RoomName == "" {
		info.RoomName = doc.ProjectName
		if info.RoomName == "" {
			info.RoomName = doc.DocumentNumber
		}
	}

	var announcer *Announcer
	if !m.cfg.DisableDiscovery {
		announcer, err = startAnnouncer(DefaultDiscoveryPort, m.cfg.AnnounceInterval, m.log)
		if err != nil {
			m.log.Warn("discovery UDP tidak aktif (join manual tetap bisa)", "error", err)
			announcer = nil
		}
	}

	r := &Room{
		info:             info,
		docID:            doc.ID,
		cfg:              m.cfg,
		log:              m.log,
		ops:              m.ops,
		sph:              m.sph,
		chat:             m.chat,
		md:               m.md,
		server:           srv,
		announcer:        announcer,
		conns:            map[string]*serverConn{},
		byID:             map[string]*Participant{hostPart.ID: &hostPart},
		assignments:      map[string][]string{},
		activeEdits:      map[string]string{},
		participantNames: map[string]string{},
		masterStatus:     map[string][]MasterStatusEntry{},
		chatLog:          []ChatPayload{},
		chatCap:          m.cfg.ActivityCap,
		stopCh:           make(chan struct{}),
		closedCh:         make(chan struct{}),
	}
	r.managerRef = m
	r.publishAnnounceLocked()
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.log.Error("presenceLoop panic", "recover", rec)
			}
		}()
		r.presenceLoop()
	}()

	m.room = r
	snap := m.sessionLocked()
	fn := m.emitFn
	m.mu.Unlock()
	if fn != nil {
		fn(snap)
	}
	m.log.Info("room kolaborasi dimulai", "room", info.RoomName, "port", info.Port)
	out := r.infoSnapshot()
	return &out, nil
}

// CloseHostedRoom menutup room yang sedang di-host dan memberi tahu semua client.
func (m *Manager) CloseHostedRoom(reason string) error {
	m.mu.Lock()
	r := m.room
	if r != nil {
		m.room = nil
	}
	fn := m.emitFn
	m.mu.Unlock()
	if r == nil {
		return services.NewConflictError("Tidak ada room yang sedang di-host.")
	}
	r.close(reason)
	m.log.Info("room kolaborasi ditutup", "reason", reason)
	if fn != nil {
		fn(UISnapshot{Mode: ModeNone, Notice: "Room \"" + r.info.RoomName + "\" ditutup.", Activities: []services.CollabActivity{}, Participants: []Participant{}})
	}
	return nil
}

// ===== CLIENT =====

// Join menghubungkan aplikasi ini sebagai client ke room host di LAN.
func (m *Manager) Join(hostIP string, port int, displayName, accessCode, roomCode string) error {
	ip := strings.TrimSpace(hostIP)
	displayName = sanitizeIdentity(displayName)
	accessCode = strings.TrimSpace(accessCode)
	if ip == "" {
		return services.NewValidationError("IP host wajib diisi (contoh 192.168.1.10).")
	}
	if port < 1024 || port > 65535 {
		return services.NewValidationError("Port harus di antara 1024 dan 65535.")
	}
	if displayName == "" {
		displayName = m.cfg.DeviceName
	}
	if accessCode == "" {
		return services.NewValidationError("Access code wajib diisi.")
	}

	m.mu.Lock()
	if m.room != nil {
		m.mu.Unlock()
		return services.NewConflictError("Aplikasi ini sedang menjadi host room. Tutup room tersebut lebih dulu.")
	}
	if m.client != nil {
		m.mu.Unlock()
		return services.NewConflictError("Sudah terhubung ke sebuah room. Keluar terlebih dulu bila ingin pindah.")
	}

	c, err := newClient(clientParams{
		addr:        net.JoinHostPort(ip, fmt.Sprintf("%d", port)),
		displayName: displayName,
		deviceName:  m.cfg.DeviceName,
		accessCode:  accessCode,
		roomCode:    strings.TrimSpace(roomCode),
		cfg:         m.cfg,
		log:         m.log,
		onStatus:    m.onClientStatus,
		onEnvelope:  m.onClientEnvelope,
		onClosed:    m.onClientClosed,
	})
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("gagal menyiapkan koneksi: %w", err)
	}

	// Pasang sesi SEBELUM start agar envelope ROOM_JOINED yang tiba saat handshake
	// tidak terbuang oleh guard m.client pada callback.
	// Penting: mu.Unlock() SEBELUM StartAndWaitReady agar onClientEnvelope tidak
	// memblokir readLoop (bila host mengirim USER_DISCONNECTED sebelum ROOM_JOINED).
	m.resetClientStateLocked()
	m.connStatus = ConnReconnecting
	m.client = c
	m.mu.Unlock()

	if err := c.StartAndWaitReady(m.cfg.DialTimeout + 2*time.Second); err != nil {
		m.mu.Lock()
		if m.client == c {
			m.client = nil
			m.resetClientStateLocked()
		}
		fn := m.emitFn
		m.mu.Unlock()
		c.stopQuiet()
		if fn != nil {
			fn(UISnapshot{Mode: ModeNone, Error: err.Error(), Activities: []services.CollabActivity{}, Participants: []Participant{}})
		}
		m.log.Error("gagal join room", "error", err)
		return err
	}

	m.mu.Lock()
	if m.client != c {
		m.mu.Unlock()
		c.stopQuiet()
		return fmt.Errorf("sesi join dibatalkan.")
	}
	m.connStatus = ConnConnected
	snap := m.sessionLocked()
	fn := m.emitFn
	m.mu.Unlock()
	if fn != nil {
		fn(snap)
	}
	m.log.Info("terhubung ke room kolaborasi", "addr", c.p.addr)
	return nil
}

// LeaveClientSession keluar dari room yang sedang diikuti.
func (m *Manager) LeaveClientSession() error {
	m.mu.Lock()
	c := m.client
	if c != nil {
		m.client = nil
	}
	fn := m.emitFn
	m.mu.Unlock()
	if c == nil {
		return services.NewConflictError("Tidak sedang join ke room mana pun.")
	}
	c.stop()
	m.mu.Lock()
	m.resetClientStateLocked()
	m.mu.Unlock()
	if fn != nil {
		fn(UISnapshot{Mode: ModeNone, Notice: "Keluar dari room.", Activities: []services.CollabActivity{}, Participants: []Participant{}})
	}
	m.log.Info("keluar dari room kolaborasi")
	return nil
}

// ===== operasi edit =====

// SendOp menjalankan satu operasi edit: host memproses lokal, client meneruskan ke host.
func (m *Manager) SendOp(op *services.OpPayload) error {
	if op == nil || strings.TrimSpace(op.Type) == "" {
		return services.NewValidationError("Operasi kosong.")
	}
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()

	switch {
	case r != nil:
		return r.applyLocal(op)
	case c != nil:
		return c.sendOp(op)
	default:
		return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
	}
}

// Session mengembalikan potret sesi saat ini (untuk halaman yang baru dibuka).
func (m *Manager) Session() UISnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionLocked()
}

// Shutdown menutup seluruh sesi & jaringan (dipanggil saat aplikasi keluar).
func (m *Manager) Shutdown(reason string) {
	m.mu.Lock()
	r, c, l := m.room, m.client, m.listener
	m.room, m.client, m.listener = nil, nil, nil
	m.mu.Unlock()

	if r != nil {
		r.close(reason)
	}
	if c != nil {
		c.stop()
	}
	if l != nil {
		l.Stop()
	}
}

// ===== discovery (lobby) =====

// StartDiscovery mulai mendengarkan broadcast room di LAN.
func (m *Manager) StartDiscovery() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listener != nil {
		return nil
	}
	l, err := startListener(DefaultDiscoveryPort, m.cfg.DiscoverTTL, m.cfg.AnnounceInterval, m.log)
	if err != nil {
		m.log.Warn("listener discovery gagal", "error", err)
		return fmt.Errorf("discovery LAN tidak dapat dijalankan (kemungkinan firewall atau port dipakai proses lain). Gunakan \"Join via IP\".")
	}
	l.onDead = m.onDiscoveryListenerDead
	m.listener = l
	return nil
}

// onDiscoveryListenerDead dipanggil saat readLoop listener mati karena error.
// Mengosongkan referensi lama lalu restart setelah jeda singkat.
func (m *Manager) onDiscoveryListenerDead() {
	m.mu.Lock()
	l := m.listener
	m.listener = nil
	m.mu.Unlock()

	if l == nil {
		return
	}

	m.log.Warn("discovery listener mati, mencoba restart dalam 1 detik...")
	time.Sleep(time.Second)

	if err := m.StartDiscovery(); err != nil {
		m.log.Warn("restart discovery listener gagal", "error", err)
	}
}

// StopDiscovery menghentikan listener lobby.
func (m *Manager) StopDiscovery() {
	m.mu.Lock()
	l := m.listener
	m.listener = nil
	m.mu.Unlock()
	if l != nil {
		l.Stop()
	}
}

// DiscoveredRooms mengembalikan daftar room yang masih dianggap hidup.
func (m *Manager) DiscoveredRooms() []DiscoveredRoom {
	m.mu.Lock()
	l := m.listener
	m.mu.Unlock()
	if l == nil {
		return []DiscoveredRoom{}
	}
	return l.Rooms()
}

// ===== builder snapshot internal (dipanggil dengan m.mu TERKUNCI) =====

func (m *Manager) sessionLocked() UISnapshot {
	if m.room != nil {
		data := m.room.eventData()
		docJSON, err := m.fetchDocJSON(m.room.docID)
		if err != nil {
			m.log.Warn("gagal memuat state dokumen untuk UI", "error", err)
		}
		info := data.Info
		return UISnapshot{
			Mode:         ModeHost,
			Room:         &info,
			Doc:          docJSON,
			Participants: info.Participants,
			Activities:   data.Activities,
			Turn:         data.Turn,
			Chat:         data.Chat,
			Unread:       m.hostUnread,
			Version:      data.Version,
			MasterStatus: data.MasterStatus,
		}
	}
	if m.client != nil {
		acts := cloneActivities(m.clientActs)
		if acts == nil {
			acts = []services.CollabActivity{}
		}
		parts := cloneParticipants(m.clientParts)
		if parts == nil {
			parts = []Participant{}
		}
		return UISnapshot{
			Mode:         ModeClient,
			Connection:   m.connStatus,
			Room:         cloneRoomInfo(m.clientRoomMeta),
			Doc:          m.clientDoc,
			Participants: parts,
			Activities:   acts,
			Turn:         m.clientTurn,
			Chat:         cloneChat(m.clientChat),
			Unread:       m.clientUnread,
			Version:      m.clientVersion,
			Error:        m.clientErr,
			Notice:       m.clientNotice,
			Incoming:     len(m.clientIncoming),
			MasterStatus: m.clientMasterStatusListLocked(),
		}
	}
	return UISnapshot{Mode: ModeNone, Activities: []services.CollabActivity{}, Participants: []Participant{}}
}

func (m *Manager) fetchDocJSON(id uint) (json.RawMessage, error) {
	doc, err := m.ops.Snapshot(id)
	if err != nil {
		return nil, err
	}
	save := services.DocToSaveInput(doc)
	b, err := json.Marshal(save)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func (m *Manager) resetClientStateLocked() {
	m.connStatus = ""
	m.clientErr = ""
	m.clientNotice = ""
	m.clientVersion = 0
	m.clientRoomMeta = nil
	m.clientDoc = nil
	m.clientParts = nil
	m.clientActs = nil
	m.clientTurn = nil
	m.clientChat = nil
	m.clientUnread = 0
	m.clientIncoming = nil
	m.clientMasterStatus = nil
}

// ===== callback client =====

// onClientStatus dipanggil client saat status koneksi berubah.
func (m *Manager) onClientStatus(status, errMsg string) {
	m.mu.Lock()
	if m.client == nil {
		m.mu.Unlock()
		return
	}
	m.connStatus = status
	m.clientErr = errMsg
	if status == ConnDisconnected {
		m.clientNotice = "Koneksi ke host terputus."
	}
	snap := m.sessionLocked()
	fn := m.emitFn
	m.mu.Unlock()
	if fn != nil {
		fn(snap)
	}
}

// onClientEnvelope memproses envelope dari host dan memperbarui cache client.
func (m *Manager) onClientEnvelope(env *Envelope) {
	if env.Type == TypeError {
		var ep ErrorPayload
		_ = json.Unmarshal(env.Payload, &ep)
		m.mu.Lock()
		if m.client != nil {
			m.clientErr = ep.Message
			snap := m.sessionLocked()
			fn := m.emitFn
			m.mu.Unlock()
			if fn != nil {
				fn(snap)
			}
		} else {
			m.mu.Unlock()
		}
		return
	}

	// Chat & master-data status masuk langsung ke snapshot UI (bukan state dokumen).
	switch env.Type {
	case TypeChatBroadcast:
		var cp ChatPayload
		_ = json.Unmarshal(env.Payload, &cp)
		m.mu.Lock()
		if m.client != nil {
			m.clientChat = append(m.clientChat, cp)
			if len(m.clientChat) > m.cfg.ActivityCap {
				m.clientChat = m.clientChat[len(m.clientChat)-m.cfg.ActivityCap:]
			}
			if cp.SenderID != "" && cp.SenderID != m.client.clientIDLocked() {
				m.clientUnread++
			}
		}
		snap := m.sessionLocked()
		fn := m.emitFn
		m.mu.Unlock()
		if fn != nil {
			fn(snap)
		}
		return
	case TypeChatHistory:
		var hp ChatHistoryPayload
		_ = json.Unmarshal(env.Payload, &hp)
		m.mu.Lock()
		if m.client != nil {
			m.clientChat = hp.Messages
			m.clientUnread = 0
		}
		snap := m.sessionLocked()
		fn := m.emitFn
		m.mu.Unlock()
		if fn != nil {
			fn(snap)
		}
		return
	case TypeMasterDataStatus:
		var mp MasterDataStatusPayload
		_ = json.Unmarshal(env.Payload, &mp)
		m.mu.Lock()
		if m.client != nil {
			var note string
			switch mp.Status {
			case MasterStatusAckInstalled:
				note = mp.TargetName + " memasang Master Data."
			case MasterStatusAckRejected:
				note = mp.TargetName + " menolak Master Data."
			default:
				note = "Master Data " + mp.TargetName + " diterima."
			}
			if note != "" {
				m.clientNotice = note
			}
			// Catat status di sisi client agar card chat client sender menampilkan status penerima.
			m.addClientMasterStatusLocked(MasterStatusEntry{
				PackageID:  mp.PackageID,
				TargetID:   mp.TargetID,
				TargetName: mp.TargetName,
				Status:     mp.Status,
				At:         time.Now().UTC(),
			})
		}
		snap := m.sessionLocked()
		fn := m.emitFn
		m.mu.Unlock()
		if fn != nil {
			fn(snap)
		}
		return
	case TypeMasterData:
		var tp MasterDataTransferPayload
		_ = json.Unmarshal(env.Payload, &tp)
		m.mu.Lock()
		if m.client != nil {
			if m.md != nil {
				_ = m.md.SaveInbox(&models.MasterInbox{
					RoomID:        env.RoomID,
					PackageID:     tp.PackageID,
					SenderID:      tp.SenderID,
					SenderName:    tp.SenderName,
					SourceVersion: tp.SourceVersion,
					Payload:       string(tp.Payload),
					Checksum:      tp.Checksum,
					Status:        models.MasterStatusPending,
					ReceivedAt:    time.Now().UTC(),
				})
			}
			cp := tp
			cp.Targets = nil
			m.clientIncoming = append(m.clientIncoming, cp)
			if len(m.clientIncoming) > m.cfg.ActivityCap {
				m.clientIncoming = m.clientIncoming[len(m.clientIncoming)-m.cfg.ActivityCap:]
			}
			if tp.SenderName != "" {
				m.clientNotice = "Master Data masuk dari " + tp.SenderName + "."
			}
		}
		snap := m.sessionLocked()
		fn := m.emitFn
		m.mu.Unlock()
		if fn != nil {
			fn(snap)
		}
		return
	}

	var sp StatePayload
	if len(env.Payload) > 0 {
		_ = json.Unmarshal(env.Payload, &sp)
	}

	m.mu.Lock()
	if m.client == nil {
		m.mu.Unlock()
		return
	}
	m.clientVersion = env.Version
	if sp.Room != nil {
		meta := sp.Room.Sanitized()
		m.clientRoomMeta = &meta
	}
	if len(sp.State) > 0 {
		m.clientDoc = append(json.RawMessage(nil), sp.State...)
	}
	if sp.Participants != nil {
		m.clientParts = cloneParticipants(sp.Participants)
	}
	if sp.Turn != nil {
		m.clientTurn = sp.Turn
	}
	switch env.Type {
	case TypeSphUpdated:
		if sp.Activity != nil {
			m.clientActs = append(m.clientActs, *sp.Activity)
			if len(m.clientActs) > m.cfg.ActivityCap {
				m.clientActs = m.clientActs[len(m.clientActs)-m.cfg.ActivityCap:]
			}
		}
	case TypeRoomJoined, TypeSyncResponse:
		m.clientActs = nil
	}
	// Setelah initial sync, client meminta riwayat chat.
	if env.Type == TypeRoomJoined || env.Type == TypeSyncRequest {
		if m.client != nil {
			_ = m.client.requestChatHistory()
		}
	}
	snap := m.sessionLocked()
	fn := m.emitFn
	m.mu.Unlock()
	if fn != nil {
		fn(snap)
	}
}

// onClientClosed dipanggil client saat sesinya ditutup permanen (stop/leave).
func (m *Manager) onClientClosed() {
	m.mu.Lock()
	active := m.client != nil
	m.mu.Unlock()
	if active {
		_ = m.LeaveClientSession()
	}
}

// ===== incoming WS (sisi host) =====

func (m *Manager) handleIncomingConn(ws *websocket.Conn) {
	m.mu.Lock()
	r := m.room
	m.mu.Unlock()
	if r == nil || r.isClosed() {
		rejectUnjoined(ws, ErrorPayload{Code: errCodeClosed, Message: "Room tidak aktif."})
		return
	}
	r.serveConn(ws)
}

// ===== ROOM =====

type Room struct {
	mu         sync.Mutex
	info       RoomInfo
	docID      uint
	version    uint64
	activities []services.CollabActivity
	chatLog    []ChatPayload
	chatCap    int

	cfg       Config
	log       *slog.Logger
	ops       *services.CollabOps
	sph       *services.SphService
	chat      ChatStore
	md        MasterDataStore
	server    *wsServer
	announcer *Announcer

	conns map[string]*serverConn  // participantID → koneksi aktif (remote)
	byID  map[string]*Participant // participantID → data peserta

	assignments      map[string][]string // participantID → []sectionID
	activeEdits      map[string]string   // sectionID → participantID
	participantNames map[string]string   // participantID → displayName (for turn state broadcast)

	masterStatus map[string][]MasterStatusEntry // packageID → status per penerima (sisi host)

	stopCh    chan struct{}
	closedCh  chan struct{}
	closeOnce sync.Once

	managerRef *Manager // diisi Manager saat membuat room (package yang sama, tanpa siklus)
}

func (r *Room) isClosed() bool {
	select {
	case <-r.closedCh:
		return true
	default:
		return false
	}
}

func (r *Room) infoSnapshot() RoomInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.infoSnapshotLocked()
}

func (r *Room) infoSnapshotLocked() RoomInfo {
	out := r.info
	out.Participants = cloneParticipants(r.info.Participants)
	out.AccessCode = r.info.AccessCode // sisi host boleh melihat access code
	return out
}

func (r *Room) activitiesLocked() []services.CollabActivity {
	out := make([]services.CollabActivity, len(r.activities))
	copy(out, r.activities)
	return out
}

// eventData mengambil salinan data untuk snapshot UI (aman dipanggil di luar lock).
func (r *Room) eventData() roomEventData {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := r.infoSnapshotLocked()
	acts := r.activitiesLocked()
	turn := r.buildTurnStateLocked()
	return roomEventData{Info: info, Activities: acts, Version: r.version, Turn: turn, Chat: cloneChat(r.chatLog), MasterStatus: r.masterStatusListLocked()}
}

type roomEventData struct {
	Info         RoomInfo
	Activities   []services.CollabActivity
	Version      uint64
	Turn         *TurnState
	Chat         []ChatPayload
	MasterStatus []MasterStatusEntry
}

// masterStatusListLocked mengembalikan salinan seluruh status master data sisi host.
// Memanggilnya dalam kondisi room.mu terkunci.
func (r *Room) masterStatusListLocked() []MasterStatusEntry {
	var out []MasterStatusEntry
	for _, entries := range r.masterStatus {
		out = append(out, entries...)
	}
	// urutkan pakai zaman agar terprediksi (terbaru dulu)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].At.After(out[j].At)
	})
	return out
}

// addHostMasterStatus mencatat status per penerima untuk satu package sisi host.
// r.mu harus terkunci oleh pemanggil.
func (r *Room) addHostMasterStatusLocked(e MasterStatusEntry) {
	if e.PackageID == "" {
		return
	}
	// jangan duplikat status yang sama untuk target yang sama
	list := r.masterStatus[e.PackageID]
	for _, existing := range list {
		if existing.TargetID != "" && existing.TargetID == e.TargetID && existing.Status == e.Status {
			return
		}
	}
	list = append(list, e)
	if len(list) > 32 {
		list = list[len(list)-32:]
	}
	r.masterStatus[e.PackageID] = list
}

// clientMasterStatusListLocked mengembalikan salinan seluruh status master data sisi client.
func (m *Manager) clientMasterStatusListLocked() []MasterStatusEntry {
	var out []MasterStatusEntry
	for _, entries := range m.clientMasterStatus {
		out = append(out, entries...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].At.After(out[j].At)
	})
	return out
}

// addClientMasterStatus mencatat status per penerima untuk satu package sisi client.
func (m *Manager) addClientMasterStatusLocked(e MasterStatusEntry) {
	if e.PackageID == "" {
		return
	}
	list := m.clientMasterStatus[e.PackageID]
	for _, existing := range list {
		if existing.TargetID != "" && existing.TargetID == e.TargetID && existing.Status == e.Status {
			return
		}
	}
	list = append(list, e)
	if len(list) > 32 {
		list = list[len(list)-32:]
	}
	m.clientMasterStatus[e.PackageID] = list
}

// notifyRoomChanged memicu pengiriman snapshot UI terbaru untuk sesi host.
// Dipanggil dari Room SETELAH room.mu dilepas.
func (m *Manager) notifyRoomChanged(r *Room) {
	m.mu.Lock()
	cur := m.room
	var snap UISnapshot
	if cur != nil && cur == r {
		snap = m.sessionLocked()
	}
	fn := m.emitFn
	m.mu.Unlock()
	if cur != nil && cur == r && fn != nil {
		fn(snap)
	}
}

// applyLocal menjalankan operasi dari host UI (aktor = nama host).
func (r *Room) applyLocal(op *services.OpPayload) error {
	actor := r.info.HostName // tidak berubah setelah room dibuat

	r.mu.Lock()
	if r.isClosed() {
		r.mu.Unlock()
		return services.NewConflictError("Room sudah ditutup.")
	}
	_, err := r.applyAndBroadcastLocked(actor, op)
	r.mu.Unlock()
	if err != nil {
		return err
	}
	r.notifyChanged()
	return nil
}

// applyAndBroadcastLocked inti penerapan operasi; room.mu HARUS terkunci.
// Mengembalikan envelope broadcast yang telah dikirim ke seluruh remote client.
func (r *Room) applyAndBroadcastLocked(actor string, op *services.OpPayload) (*Envelope, error) {
	doc, act, err := r.ops.Apply(r.docID, actor, op)
	if err != nil {
		r.log.Warn("operasi kolaborasi ditolak", "op", op.Type, "actor", actor, "error", err)
		return nil, err
	}
	r.version++
	act.Actor = actor
	r.activities = append(r.activities, *act)
	if len(r.activities) > r.cfg.ActivityCap {
		r.activities = r.activities[len(r.activities)-r.cfg.ActivityCap:]
	}

	save := services.DocToSaveInput(doc)
	state, err := json.Marshal(save)
	if err != nil {
		r.log.Error("gagal serialisasi state", "error", err)
		return nil, fmt.Errorf("gagal menyiapkan state")
	}
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		State:        state,
		Activity:     act,
		Participants: cloneParticipants(r.info.Participants),
		Turn:         r.buildTurnStateLocked(),
	}
	env := envelopeWith(TypeSphUpdated, r.info.RoomID, "", r.version, payload)
	r.broadcastLocked(env)
	return env, nil
}

// sanitizedInfoLocked menyalin info room tanpa access code untuk disebarkan.
func (r *Room) sanitizedInfoLocked() *RoomInfo {
	info := r.infoSnapshotLocked()
	info.AccessCode = ""
	return &info
}

// broadcastLocked mengirim envelope ke semua koneksi remote (non-blocking).
func (r *Room) broadcastLocked(env *Envelope) {
	for id, c := range r.conns {
		if !c.deliver(env) {
			r.log.Warn("client lambat diputus", "participant", id)
			c.shutdown()
			delete(r.conns, id)
		}
	}
}

func (r *Room) sendError(c *serverConn, code, message string) {
	c.deliver(envelopeWith(TypeError, "", "", 0, ErrorPayload{Code: code, Message: message}))
}

// publishAnnounceLocked memperbarui paket UDP discovery dengan jumlah pengguna kini.
func (r *Room) publishAnnounceLocked() {
	if r.announcer == nil {
		return
	}
	r.announcer.Set(announcePacket{
		RoomID:         r.info.RoomID,
		RoomName:       r.info.RoomName,
		DocumentNumber: r.info.DocumentNumber,
		ProjectName:    r.info.ProjectName,
		HostName:       r.info.HostName,
		WSPort:         r.info.Port,
		Users:          len(r.info.Participants),
	})
}

// serveConn menangani satu koneksi client dari upgrade hingga tutup.
func (r *Room) serveConn(ws *websocket.Conn) {
	conn := newServerConn(ws)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.log.Error("writePump panic", "recover", rec)
			}
		}()
		conn.writePump(r.log)
	}()

	ws.SetReadDeadline(time.Now().Add(r.cfg.JoinWait))
	var first Envelope
	if err := ws.ReadJSON(&first); err != nil {
		r.log.Warn("serveConn: gagal baca pesan pertama", "error", err)
		conn.shutdown()
		return
	}
	r.log.Info("serveConn: pesan pertama diterima", "type", first.Type, "remote", ws.RemoteAddr().String())
	if first.Type != TypeJoinRequest {
		r.log.Warn("serveConn: pesan pertama bukan JOIN_REQUEST", "type", first.Type)
		r.sendError(conn, errCodeNotJoined, "Pesan pertama harus JOIN_REQUEST.")
		conn.shutdown()
		return
	}
	var jr JoinRequest
	if len(first.Payload) > 0 {
		if err := json.Unmarshal(first.Payload, &jr); err != nil {
			r.log.Warn("serveConn: JOIN_REQUEST payload tidak valid", "error", err)
			r.sendError(conn, errCodeNotJoined, "JOIN_REQUEST tidak valid.")
			conn.shutdown()
			return
		}
	}
	r.log.Info("serveConn: JOIN_REQUEST diterima", "name", jr.DisplayName, "device", jr.DeviceName, "hasClientID", jr.ClientID != "")

	p, replaced, joined := r.authenticate(&jr, conn)
	if !joined {
		r.log.Warn("serveConn: authenticate gagal", "name", jr.DisplayName)
		conn.shutdown()
		return
	}
	conn.participantID = p.ID
	if replaced != nil {
		replaced.shutdown()
	}

	r.log.Info("serveConn: mengirim initial state", "participant", p.DisplayName, "docID", r.docID)
	r.sendInitialState(conn)
	r.broadcastPresence(TypeUserConnected, p)
	r.notifyChanged()
	r.log.Info("serveConn: client terhubung", "name", p.DisplayName, "id", p.ID)

	for {
		ws.SetReadDeadline(time.Now().Add(r.cfg.ReadWait))
		var e Envelope
		if err := ws.ReadJSON(&e); err != nil {
			break
		}
		switch e.Type {
		case TypePing:
			r.touch(p.ID)
			conn.deliver(envelopeWith(TypePong, r.info.RoomID, "", 0, nil))
		case TypeOpRequest:
			if p.Role != RoleHost {
				r.sendError(conn, errCodeOp, "Hanya host yang dapat mengirim operasi edit langsung.")
				continue
			}
			var op services.OpPayload
			if err := json.Unmarshal(e.Payload, &op); err != nil {
				r.sendError(conn, errCodeOp, "Operasi tidak valid.")
				continue
			}
			r.applyRemote(conn, p, &op)
		case TypeAssignTurns:
			var ta map[string][]string
			if err := json.Unmarshal(e.Payload, &ta); err != nil {
				r.sendError(conn, errCodeOp, "Payload assign turns tidak valid.")
				continue
			}
			r.handleAssignTurns(ta)
		case TypeRequestEdit:
			var sectionID string
			if err := json.Unmarshal(e.Payload, &sectionID); err != nil {
				r.sendError(conn, errCodeOp, "Payload request edit tidak valid.")
				continue
			}
			r.handleRequestEdit(p, sectionID)
		case TypeReleaseEdit:
			var sectionID string
			if err := json.Unmarshal(e.Payload, &sectionID); err != nil {
				r.sendError(conn, errCodeOp, "Payload release edit tidak valid.")
				continue
			}
			r.handleReleaseEdit(p, sectionID)
		case TypeSyncPush:
			var input services.SphSaveInput
			if err := json.Unmarshal(e.Payload, &input); err != nil {
				r.sendError(conn, errCodeOp, "Payload sync push tidak valid.")
				continue
			}
			r.handleSyncPush(p, &input)
		case TypeSyncRequest:
			r.sendInitialState(conn)
		case TypeChatMessage:
			var cp ChatPayload
			if err := json.Unmarshal(e.Payload, &cp); err != nil {
				r.sendError(conn, errCodeOp, "Payload chat tidak valid.")
				continue
			}
			r.handleIncomingChat(p, &cp)
		case TypeChatHistoryRequest:
			r.handleChatHistoryRequest(conn)
		case TypeMasterData:
			var tp MasterDataTransferPayload
			if err := json.Unmarshal(e.Payload, &tp); err != nil {
				r.sendError(conn, errCodeOp, "Payload master data tidak valid.")
				continue
			}
			r.relayMasterData(p, &tp)
		case TypeMasterDataAck:
			var ap MasterDataAckPayload
			if err := json.Unmarshal(e.Payload, &ap); err != nil {
				r.sendError(conn, errCodeOp, "Payload master data ack tidak valid.")
				continue
			}
			r.handleMasterDataAck(conn, p, &ap)
		case TypeLeave:
			r.dropParticipant(p.ID, conn, false)
			conn.shutdown()
			return
		default:
			r.sendError(conn, errCodeNotJoined, "Tipe pesan tidak dikenal: "+e.Type)
		}
	}

	r.dropParticipant(p.ID, conn, true)
	conn.shutdown()
}

// authenticate memeriksa access/room code lalu mendaftarkan (atau menyambung ulang) peserta.
// Mengembalikan peserta, koneksi lama yang digantikan (bila reconnect), dan status sukses.
func (r *Room) authenticate(jr *JoinRequest, conn *serverConn) (*Participant, *serverConn, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed() {
		r.log.Warn("authenticate: room sudah ditutup")
		r.sendError(conn, errCodeClosed, "Room sudah ditutup.")
		return nil, nil, false
	}
	if !equalConstTime(strings.TrimSpace(jr.AccessCode), r.info.AccessCode) {
		r.log.Warn("authenticate: access code salah", "device", jr.DeviceName, "name", jr.DisplayName, "input_len", len(jr.AccessCode), "expected_len", len(r.info.AccessCode))
		r.sendError(conn, errCodeAuth, "Access code salah. Minta kode kepada host room.")
		return nil, nil, false
	}
	if rc := strings.TrimSpace(jr.RoomCode); rc != "" && !strings.EqualFold(rc, strings.TrimSpace(r.info.RoomCode)) {
		r.log.Warn("authenticate: room code salah", "input", rc, "expected", r.info.RoomCode)
		r.sendError(conn, errCodeAuth, "Room code salah.")
		return nil, nil, false
	}

	now := time.Now()
	if cid := strings.TrimSpace(jr.ClientID); cid != "" {
		if p, ok := r.byID[cid]; ok {
			p.LastSeen = now
			old := r.conns[cid]
			r.conns[cid] = conn
			r.participantNames[cid] = p.DisplayName
			r.ensureParticipantListed(p)
			r.publishAnnounceLocked()
			return p, old, true
		}
		// identitas lama sudah dibersihkan → daftar ulang memakai ID yang sama bila bebas
		name := sanitizeIdentity(jr.DisplayName)
		if name == "" {
			name = "Tamu"
		}
		device := sanitizeIdentity(jr.DeviceName)
		if device == "" {
			device = r.cfg.DeviceName
		}
		p := &Participant{ID: cid, DisplayName: name, DeviceName: device, Role: RoleEditor, JoinedAt: now, LastSeen: now}
		r.byID[cid] = p
		r.participantNames[cid] = p.DisplayName
		r.conns[cid] = conn
		r.ensureParticipantListed(p)
		r.publishAnnounceLocked()
		return p, nil, true
	}

	name := sanitizeIdentity(jr.DisplayName)
	if name == "" {
		name = "Tamu"
	}
	device := sanitizeIdentity(jr.DeviceName)
	if device == "" {
		device = r.cfg.DeviceName
	}
	p := &Participant{
		ID:          uuid.NewString(),
		DisplayName: name,
		DeviceName:  device,
		Role:        RoleEditor,
		JoinedAt:    now,
		LastSeen:    now,
	}
	r.byID[p.ID] = p
	r.participantNames[p.ID] = p.DisplayName
	r.conns[p.ID] = conn
	r.ensureParticipantListed(p)
	r.publishAnnounceLocked()
	r.log.Info("authenticate: berhasil", "name", p.DisplayName, "role", p.Role, "id", p.ID)
	return p, nil, true
}

// ensureParticipantListed menambahkan participant ke r.info.Participants jika belum ada.
// Dipanggil saat client join atau reconnect.
func (r *Room) ensureParticipantListed(p *Participant) {
	for i := range r.info.Participants {
		if r.info.Participants[i].ID == p.ID {
			return
		}
	}
	r.info.Participants = append(r.info.Participants, *p)
}

// sendInitialState mengirim ROOM_JOINED berisi state dokumen penuh ke satu client.
// Dipanggil tanpa memegang lock saat membaca DB agar operasi lain tidak tertahan;
// versi yang dicantumkan diambil sesaat sebelum pengiriman (state selalu cukup baru).
func (r *Room) sendInitialState(c *serverConn) {
	doc, err := r.ops.Snapshot(r.docID)
	if err != nil {
		r.log.Error("sendInitialState: gagal memuat state", "error", err, "docID", r.docID)
		r.sendError(c, errCodeInternal, "Gagal menyiapkan state dokumen.")
		return
	}
	save := services.DocToSaveInput(doc)
	state, err := json.Marshal(save)
	if err != nil {
		r.log.Error("sendInitialState: gagal marshal state", "error", err)
		r.sendError(c, errCodeInternal, "Gagal menyiapkan state dokumen.")
		return
	}

	r.mu.Lock()
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		State:        state,
		Participants: cloneParticipants(r.info.Participants),
		Turn:         r.buildTurnStateLocked(),
	}
	env := envelopeWith(TypeRoomJoined, r.info.RoomID, c.participantID, r.version, payload)
	ok := c.deliver(env)
	r.mu.Unlock()
	if !ok {
		r.log.Warn("sendInitialState: deliver gagal (buffer penuh?)", "participant", c.participantID)
		c.shutdown()
		return
	}
	r.log.Info("sendInitialState: ROOM_JOINED terkirim", "participant", c.participantID, "stateSize", len(state), "version", r.version)
}

// applyRemote menjalankan operasi dari client remote; error dikirim balik hanya ke pengirim.
func (r *Room) applyRemote(c *serverConn, p *Participant, op *services.OpPayload) {
	r.mu.Lock()
	if r.isClosed() {
		r.mu.Unlock()
		r.sendError(c, errCodeClosed, "Room sudah ditutup.")
		return
	}
	_, err := r.applyAndBroadcastLocked(p.DisplayName, op)
	r.mu.Unlock()

	if err != nil {
		if services.IsFriendly(err) {
			r.sendError(c, errCodeOp, err.Error())
		} else {
			r.sendError(c, errCodeInternal, "Operasi gagal diproses di host.")
		}
		return
	}
	r.notifyChanged()
}

// touch menyegarkan waktu terakhir peserta terlihat (heartbeat PING).
func (r *Room) touch(participantID string) {
	r.mu.Lock()
	if p, ok := r.byID[participantID]; ok {
		p.LastSeen = time.Now()
	}
	r.mu.Unlock()
}

// hostParticipantLocked mengembalikan participant berperan HOST. Dipanggil dengan
// room.mu TERKUNCI (atau saat masih safe, room tidak sedang berubah).
func (r *Room) hostParticipantLocked() *Participant {
	for _, p := range r.byID {
		if p.Role == RoleHost {
			return p
		}
	}
	return nil
}

// broadcastPresence menyiarkan perubahan kehadiran peserta.
func (r *Room) broadcastPresence(t string, p *Participant) {
	r.mu.Lock()
	if r.isClosed() {
		r.mu.Unlock()
		return
	}
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		Participants: cloneParticipants(r.info.Participants),
	}
	env := envelopeWith(t, r.info.RoomID, p.ID, r.version, payload)
	r.broadcastLocked(env)
	r.publishAnnounceLocked()
	r.mu.Unlock()
}

// dropParticipant menghapus peserta; onlyIfCurrentConn menjaga reconnect tidak
// menghapus peserta yang sudah tersambung ulang lewat koneksi baru.
func (r *Room) dropParticipant(participantID string, conn *serverConn, onlyIfCurrentConn bool) {
	r.mu.Lock()
	if onlyIfCurrentConn {
		if cur, ok := r.conns[participantID]; ok && cur != conn {
			r.mu.Unlock()
			return
		}
	}
	p, ok := r.byID[participantID]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.byID, participantID)
	delete(r.conns, participantID)
	delete(r.participantNames, participantID)
	for section, editor := range r.activeEdits {
		if editor == participantID {
			delete(r.activeEdits, section)
		}
	}
	delete(r.assignments, participantID)
	for i, ep := range r.info.Participants {
		if ep.ID == participantID {
			r.info.Participants = append(r.info.Participants[:i], r.info.Participants[i+1:]...)
			break
		}
	}
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		Participants: cloneParticipants(r.info.Participants),
	}
	env := envelopeWith(TypeUserDisonnected, r.info.RoomID, participantID, r.version, payload)
	r.broadcastLocked(env)
	r.publishAnnounceLocked()
	r.mu.Unlock()

	r.log.Info("peserta keluar", "name", p.DisplayName)
	r.notifyChanged()
}

// presenceLoop membersihkan peserta yang hilang tanpa pamit (koneksi mati total).
func (r *Room) presenceLoop() {
	interval := r.cfg.PresenceTTL / 2
	if interval < 500*time.Millisecond {
		interval = 500 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			stale := func() []*Participant {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.isClosed() {
					return nil
				}
				now := time.Now()
				var out []*Participant
				for id, p := range r.byID {
					if p.Role == RoleHost {
						continue
					}
					if _, live := r.conns[id]; live {
						continue
					}
					if now.Sub(p.LastSeen) > r.cfg.PresenceTTL {
						out = append(out, p)
					}
				}
				for _, p := range out {
					delete(r.byID, p.ID)
					delete(r.conns, p.ID)
				}
				if len(out) > 0 {
					payload := StatePayload{
						Room:         r.sanitizedInfoLocked(),
						Participants: cloneParticipants(r.info.Participants),
					}
					env := envelopeWith(TypeUserDisonnected, r.info.RoomID, "", r.version, payload)
					r.broadcastLocked(env)
					r.publishAnnounceLocked()
				}
				return out
			}()
			if len(stale) > 0 {
				for _, p := range stale {
					r.log.Info("peserta hilang (timeout)", "name", p.DisplayName)
				}
				r.notifyChanged()
			}
		}
	}
}

// close menutup room: broadcast ROOM_CLOSED, matikan server & announcer.
func (r *Room) close(reason string) {
	r.closeOnce.Do(func() {
		env := envelopeWith(TypeRoomClosed, "", "", 0, ClosedPayload{Reason: reason})

		r.mu.Lock()
		for id, c := range r.conns {
			c.deliver(env)
			c.shutdown()
			delete(r.conns, id)
		}
		server := r.server
		announcer := r.announcer
		r.info.Status = RoomStatusClosed
		r.mu.Unlock()

		close(r.stopCh)
		close(r.closedCh)
		if server != nil {
			server.Close()
		}
		if announcer != nil {
			announcer.Stop()
		}
		r.log.Info("room ditutup", "reason", reason)
	})
}

func (r *Room) handleAssignTurns(assignments map[string][]string) {
	r.mu.Lock()
	r.assignments = assignments
	r.activeEdits = map[string]string{}
	r.version++
	tState := r.buildTurnStateLocked()
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		Participants: cloneParticipants(r.info.Participants),
		Turn:         tState,
	}
	env := envelopeWith(TypeTurnsUpdated, r.info.RoomID, "", r.version, payload)
	r.broadcastLocked(env)
	r.mu.Unlock()
	r.notifyChanged()
}

func (r *Room) handleRequestEdit(p *Participant, sectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed() {
		r.sendErrorToParticipant(p.ID, "Room sudah ditutup.")
		return
	}

	validSections := map[string]bool{"header": true, "items": true, "subitems": true}
	if !validSections[sectionID] {
		r.sendErrorToParticipant(p.ID, "Section tidak valid: "+sectionID)
		return
	}

	allowed := r.assignments[p.ID]
	hasAccess := false
	for _, s := range allowed {
		if s == sectionID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		r.sendErrorToParticipant(p.ID, "Anda tidak memiliki akses ke section ini.")
		return
	}

	if editor, ok := r.activeEdits[sectionID]; ok && editor != p.ID {
		r.sendErrorToParticipant(p.ID, "Section sedang diedit oleh "+r.participantNames[editor]+".")
		return
	}

	r.activeEdits[sectionID] = p.ID
	r.version++

	tState := r.buildTurnStateLocked()
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		Participants: cloneParticipants(r.info.Participants),
		Turn:         tState,
	}
	env := envelopeWith(TypeTurnsUpdated, r.info.RoomID, "", r.version, payload)
	r.broadcastLocked(env)
	r.notifyChanged()
}

func (r *Room) handleReleaseEdit(p *Participant, sectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed() {
		return
	}

	editor, ok := r.activeEdits[sectionID]
	if !ok || editor != p.ID {
		return
	}

	delete(r.activeEdits, sectionID)
	r.version++

	tState := r.buildTurnStateLocked()
	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		Participants: cloneParticipants(r.info.Participants),
		Turn:         tState,
	}
	env := envelopeWith(TypeTurnsUpdated, r.info.RoomID, "", r.version, payload)
	r.broadcastLocked(env)
	r.notifyChanged()
}

func (r *Room) handleSyncPush(p *Participant, input *services.SphSaveInput) {
	r.mu.Lock()
	if r.isClosed() {
		r.mu.Unlock()
		r.sendErrorToParticipant(p.ID, "Room sudah ditutup.")
		return
	}

	doc, err := r.sph.ApplySave(r.docID, input, p.DisplayName)
	if err != nil {
		r.mu.Unlock()
		r.sendErrorToParticipant(p.ID, "Gagal menerapkan perubahan: "+err.Error())
		return
	}

	r.version++
	save := services.DocToSaveInput(doc)
	state, err := json.Marshal(save)
	if err != nil {
		r.mu.Unlock()
		r.sendErrorToParticipant(p.ID, "Gagal memproses dokumen.")
		return
	}

	act := &services.CollabActivity{
		Actor:   p.DisplayName,
		Action:  "SYNC",
		Summary: "menyinkronkan perubahan.",
	}
	r.activities = append(r.activities, *act)
	if len(r.activities) > r.cfg.ActivityCap {
		r.activities = r.activities[len(r.activities)-r.cfg.ActivityCap:]
	}

	payload := StatePayload{
		Room:         r.sanitizedInfoLocked(),
		State:        state,
		Activity:     act,
		Participants: cloneParticipants(r.info.Participants),
		Turn:         r.buildTurnStateLocked(),
	}
	env := envelopeWith(TypeSphUpdated, r.info.RoomID, "", r.version, payload)
	r.broadcastLocked(env)
	r.mu.Unlock()
	r.notifyChanged()
}

func (r *Room) buildTurnStateLocked() *TurnState {
	return &TurnState{
		Assignments: r.assignments,
		ActiveEdits: r.activeEdits,
	}
}

func (r *Room) sendErrorToParticipant(participantID, message string) {
	if c, ok := r.conns[participantID]; ok {
		c.deliver(envelopeWith(TypeError, "", "", 0, ErrorPayload{Code: errCodeOp, Message: message}))
	}
}

// ===== Chat (host) =====

// handleIncomingChat menerima chat dari client/host, mempersist di host, lalu
// broadcast ke seluruh member (termasuk echo ke pengirim).
func (r *Room) handleIncomingChat(p *Participant, cp *ChatPayload) {
	if cp == nil {
		return
	}
	now := time.Now()
	msg := ChatMessage{
		RoomID:      r.info.RoomID,
		MessageID:   cp.MessageID,
		SenderID:    p.ID,
		SenderName:  p.DisplayName,
		MessageType: cp.MessageType,
		Content:     cp.Content,
		Status:      ChatStatusDelivered,
		CreatedAt:   now,
	}
	if msg.MessageID == "" {
		msg.MessageID = uuid.NewString()
	}
	if msg.MessageType == "" {
		msg.MessageType = ChatTypeText
	}

	if r.chat != nil {
		if err := r.chat.SaveChat(msg); err != nil {
			r.log.Warn("gagal menyimpan chat di host", "error", err)
		}
	}

	out := ChatPayload{
		MessageID:   msg.MessageID,
		RoomID:      msg.RoomID,
		SenderID:    msg.SenderID,
		SenderName:  msg.SenderName,
		MessageType: msg.MessageType,
		Content:     msg.Content,
		Status:      msg.Status,
		RefPackage:  cp.RefPackage,
		RefMeta:     cp.RefMeta,
		CreatedAt:   msg.CreatedAt,
	}

	r.mu.Lock()
	if r.isClosed() {
		r.mu.Unlock()
		return
	}
	remote := true
	for _, pp := range r.byID {
		if pp.Role == RoleHost && pp.ID == p.ID {
			remote = false
			break
		}
	}
	r.chatLog = append(r.chatLog, out)
	if len(r.chatLog) > r.chatCap {
		r.chatLog = r.chatLog[len(r.chatLog)-r.chatCap:]
	}
	env := envelopeWith(TypeChatBroadcast, r.info.RoomID, "", 0, out)
	r.broadcastLocked(env)
	r.mu.Unlock()
	if remote {
		if mgr := r.managerRef; mgr != nil {
			mgr.mu.Lock()
			mgr.hostUnread++
			snap := mgr.sessionLocked()
			fn := mgr.emitFn
			mgr.mu.Unlock()
			if fn != nil {
				fn(snap)
			}
		}
	}
	r.notifyChanged()
}

// handleChatHistoryRequest mengirim riwayat chat terakhir ke client peminta.
func (r *Room) handleChatHistoryRequest(conn *serverConn) {
	if r.chat == nil {
		conn.deliver(envelopeWith(TypeChatHistory, r.info.RoomID, "", 0, ChatHistoryPayload{}))
		return
	}
	msgs, err := r.chat.History(r.info.RoomID, 200)
	if err != nil {
		r.log.Warn("gagal memuat riwayat chat", "error", err)
		msgs = nil
	}
	out := make([]ChatPayload, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, ChatPayload{
			MessageID:   m.MessageID,
			RoomID:      m.RoomID,
			SenderID:    m.SenderID,
			SenderName:  m.SenderName,
			MessageType: m.MessageType,
			Content:     m.Content,
			Status:      m.Status,
			CreatedAt:   m.CreatedAt,
		})
	}
	env := envelopeWith(TypeChatHistory, r.info.RoomID, "", 0, ChatHistoryPayload{Messages: out})
	conn.deliver(env)
}

// ===== Master Data (host relay) =====

// handleIncomingMasterData menerima MasterDataPackage dari client, memvalidasi,
// lalu meneruskan ke target (satu/banyak/semua) atau diproses sendiri bila target=host.
// Implementasi relay penuh dirangkai di lapisan Manager/app; di Room ini kita menangani
// notifikasi ACK status agar pengirim mendapat status penerimaan.
//
// Catatan: pengiriman MasterDataPackage dilakukan lewat Manager (bukan raw serveConn)
// karena butuh akses DB lokal host (inbox) untuk target host. Di sini hanya ACK status.
func (r *Room) handleMasterDataAck(conn *serverConn, p *Participant, ap *MasterDataAckPayload) {
	if ap == nil || ap.PackageID == "" {
		r.sendError(conn, errCodeOp, "Payload master data ack tidak valid.")
		return
	}
	status := MasterDataAckPayload{
		PackageID:  ap.PackageID,
		Status:     ap.Status,
		Message:    ap.Message,
		AckedAt:    time.Now(),
		TargetID:   p.ID,
		TargetName: p.DisplayName,
	}
	// Jika host adalah pengirim, mutakhirkan status pengiriman lokal host.
	if r.md != nil {
		_ = r.md.SetSentStatus(ap.PackageID, ap.Status)
	}
	// Catat status di sisi host agar card chat host sender menampilkan status penerima.
	r.mu.Lock()
	if !r.isClosed() {
		r.addHostMasterStatusLocked(MasterStatusEntry{
			PackageID:  ap.PackageID,
			TargetID:   p.ID,
			TargetName: p.DisplayName,
			Status:     ap.Status,
			At:         time.Now().UTC(),
		})
	}
	env := envelopeWith(TypeMasterDataStatus, r.info.RoomID, "", 0, status)
	if !r.isClosed() {
		r.broadcastLocked(env)
	}
	r.mu.Unlock()
	r.notifyChanged()
}

// hostSendMasterData dipanggil saat HOST mengirim Master Data (bukan relay dari client):
// mencatat sent, menambah pesan chat, lalu meneruskan ke tiap target.
func (r *Room) hostSendMasterData(tp *MasterDataTransferPayload, targets []string) error {
	r.mu.Lock()
	host := r.hostParticipantLocked()
	r.mu.Unlock()
	if host == nil {
		return services.NewConflictError("Host room tidak ditemukan.")
	}
	tp.SenderID = host.ID
	tp.SenderName = host.DisplayName
	tp.Targets = nil

	if r.md != nil {
		recips, _ := json.Marshal(targets)
		_ = r.md.SaveSent(&models.MasterSent{
			RoomID:        r.info.RoomID,
			PackageID:     tp.PackageID,
			Payload:       string(tp.Payload),
			Checksum:      tp.Checksum,
			SourceVersion: tp.SourceVersion,
			Recipients:    string(recips),
			Status:        models.MasterStatusPending,
			SentAt:        time.Now().UTC(),
		})
	}
	r.addMasterChat(host.ID, host.DisplayName, tp)
	return r.dispatchMasterData(tp, targets, host.ID)
}

// relayMasterData menerima paket dari client (serveConn TypeMasterData) lalu
// meneruskannya ke target. Sent dicatat di sisi pengirim (client) — di sini host
// hanya meneruskan & memproses target host via inbox.
func (r *Room) relayMasterData(sender *Participant, tp *MasterDataTransferPayload) {
	if sender == nil || tp == nil || tp.PackageID == "" || len(tp.Payload) == 0 {
		return
	}
	targets := tp.Targets
	tp.Targets = nil
	r.mu.Lock()
	host := r.hostParticipantLocked()
	r.mu.Unlock()
	if host != nil {
		r.addMasterChat(sender.ID, sender.DisplayName, tp)
	}
	_ = r.dispatchMasterData(tp, targets, "")
	_ = host
}

// dispatchMasterData meneruskan tp ke tiap target: target==host → inbox host;
// selain itu → deliver ke koneksi remote. hostID menandai identitas host.
func (r *Room) dispatchMasterData(tp *MasterDataTransferPayload, targets []string, hostID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed() {
		return services.NewConflictError("Room sudah ditutup.")
	}
	if hostID == "" {
		host := r.hostParticipantLocked()
		if host != nil {
			hostID = host.ID
		}
	}
	env := envelopeWith(TypeMasterData, r.info.RoomID, "", 0, tp)
	seen := map[string]bool{}
	for _, tid := range targets {
		if tid == "" || seen[tid] {
			continue
		}
		seen[tid] = true
		if tid == hostID {
			r.saveHostInboxLocked(tp)
			continue
		}
		if c, ok := r.conns[tid]; ok {
			c.deliver(env)
		}
	}
	return nil
}

// saveHostInboxLocked menyimpan paket masuk ke inbox host (target host).
func (r *Room) saveHostInboxLocked(tp *MasterDataTransferPayload) {
	if r.md == nil || tp == nil {
		return
	}
	_ = r.md.SaveInbox(&models.MasterInbox{
		RoomID:        r.info.RoomID,
		PackageID:     tp.PackageID,
		SenderID:      tp.SenderID,
		SenderName:    tp.SenderName,
		SourceVersion: tp.SourceVersion,
		Payload:       string(tp.Payload),
		Checksum:      tp.Checksum,
		Status:        models.MasterStatusPending,
		ReceivedAt:    time.Now().UTC(),
	})
}

// addMasterChat membuat pesan chat bertipe master_data lalu menyiarkan, agar
// seluruh member (termasuk pengirim) melihat card transfer di panel chat.
func (r *Room) addMasterChat(senderID, senderName string, tp *MasterDataTransferPayload) {
	msg := ChatMessage{
		RoomID:      r.info.RoomID,
		MessageID:   uuid.NewString(),
		SenderID:    senderID,
		SenderName:  senderName,
		MessageType: ChatTypeMasterData,
		Content:     tp.Summary,
		Status:      ChatStatusDelivered,
		CreatedAt:   time.Now().UTC(),
	}
	if r.chat != nil {
		_ = r.chat.SaveChat(msg)
	}
	out := ChatPayload{
		MessageID:   msg.MessageID,
		RoomID:      msg.RoomID,
		SenderID:    msg.SenderID,
		SenderName:  msg.SenderName,
		MessageType: msg.MessageType,
		Content:     msg.Content,
		Status:      msg.Status,
		RefPackage:  tp.PackageID,
		RefMeta:     tp.Summary,
		CreatedAt:   msg.CreatedAt,
	}
	r.mu.Lock()
	if !r.isClosed() {
		r.chatLog = append(r.chatLog, out)
		if len(r.chatLog) > r.chatCap {
			r.chatLog = r.chatLog[len(r.chatLog)-r.chatCap:]
		}
		env := envelopeWith(TypeChatBroadcast, r.info.RoomID, "", 0, out)
		r.broadcastLocked(env)
	}
	r.mu.Unlock()
	r.notifyChanged()
}

// Manager.SendMasterData mengirim Master Data ke sesi aktif.
// targets berisi participant IDs tujuan; bisa memuat ID host untuk dikirim ke dirinya.
func (m *Manager) SendMasterData(pkg *MasterDataPackage, targets []string) error {
	if pkg == nil {
		return services.NewValidationError("Package Master Data kosong.")
	}
	raw, err := pkg.Serialize()
	if err != nil {
		return fmt.Errorf("gagal serialisasi package: %w", err)
	}
	tp := &MasterDataTransferPayload{
		PackageID:     pkg.Metadata.PackageID,
		SenderID:      pkg.Metadata.SenderID,
		SenderName:    pkg.Metadata.SenderName,
		SourceVersion: pkg.Metadata.SourceVersion,
		Checksum:      pkg.Checksum,
		Title:         summarizePackage(pkg),
		Summary:       summarizePackage(pkg),
		Targets:       targets,
		Payload:       raw,
	}
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	switch {
	case r != nil:
		return r.hostSendMasterData(tp, targets)
	case c != nil:
		return c.sendMasterData(tp)
	default:
		return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
	}
}

func summarizePackage(pkg *MasterDataPackage) string {
	if pkg == nil {
		return "Master Data"
	}
	n := len(pkg.Data.Categories) + len(pkg.Data.WorkItems) + len(pkg.Data.WorkSubItems) + len(pkg.Data.Materials)
	return fmt.Sprintf("Master Data (%d item)", n)
}

// CurrentIdentity mengembalikan ID & nama peserta sesi aktif (host atau client),
// untuk mengisi identitas pengirim Master Data.
func (m *Manager) CurrentIdentity() (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.room != nil {
		m.room.mu.Lock()
		defer m.room.mu.Unlock()
		if h := m.room.hostParticipantLocked(); h != nil {
			return h.ID, h.DisplayName
		}
	}
	if m.client != nil {
		return m.client.clientIDLocked(), m.client.p.displayName
	}
	return "", ""
}

// CurrentRoomID mengembalikan ID Room sesi aktif.
func (m *Manager) CurrentRoomID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.room != nil {
		return m.room.info.RoomID
	}
	if m.client != nil && m.clientRoomMeta != nil {
		return m.clientRoomMeta.RoomID
	}
	return ""
}

// AcknowledgeMasterData melaporkan status pemasangan paket masuk (INSTALLED/
// REJECTED/FAILED) ke pengirim. Untuk client dikirim ke host; untuk host (server)
// cukup diperbarui lokal lewat store.
func (m *Manager) AcknowledgeMasterData(packageID, status, message string) error {
	if packageID == "" {
		return services.NewValidationError("Package ID wajib diisi.")
	}
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	switch {
	case c != nil:
		ap := &MasterDataAckPayload{PackageID: packageID, Status: status, Message: message, AckedAt: time.Now().UTC()}
		return c.sendMasterDataAck(ap)
	case r != nil:
		if r.md != nil {
			if err := r.md.SetSentStatus(packageID, status); err != nil {
				return err
			}
		}
		return nil
	default:
		return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
	}
}

// notifyChanged memberi tahu UI setelah lock dilepas.
func (r *Room) notifyChanged() {
	if mgr := r.managerRef; mgr != nil {
		mgr.notifyRoomChanged(r)
	}
}

func (m *Manager) AssignTurns(assignments map[string][]string) error {
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	if r != nil {
		r.handleAssignTurns(assignments)
		return nil
	}
	if c != nil {
		return c.sendAssignTurns(assignments)
	}
	return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
}

func (m *Manager) RequestEdit(sectionID string) error {
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	if r != nil {
		hostID := ""
		r.mu.Lock()
		for id, p := range r.byID {
			if p.Role == RoleHost {
				hostID = id
				break
			}
		}
		hostPart := r.byID[hostID]
		r.mu.Unlock()
		if hostPart != nil {
			r.handleRequestEdit(hostPart, sectionID)
		}
		return nil
	}
	if c != nil {
		return c.sendRequestEdit(sectionID)
	}
	return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
}

func (m *Manager) ReleaseEdit(sectionID string) error {
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	if r != nil {
		hostID := ""
		r.mu.Lock()
		for id, p := range r.byID {
			if p.Role == RoleHost {
				hostID = id
				break
			}
		}
		hostPart := r.byID[hostID]
		r.mu.Unlock()
		if hostPart != nil {
			r.handleReleaseEdit(hostPart, sectionID)
		}
		return nil
	}
	if c != nil {
		return c.sendReleaseEdit(sectionID)
	}
	return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
}

func (m *Manager) SyncPush(input *services.SphSaveInput) error {
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()
	if r != nil {
		r.mu.Lock()
		docID := r.docID
		hostName := r.info.HostName
		r.mu.Unlock()
		doc, err := m.sph.ApplySave(docID, input, hostName)
		if err != nil {
			return err
		}
		r.mu.Lock()
		r.version++
		save := services.DocToSaveInput(doc)
		state, _ := json.Marshal(save)
		act := &services.CollabActivity{
			Actor:   r.info.HostName,
			Action:  "SYNC",
			Summary: "menyinkronkan perubahan.",
		}
		r.activities = append(r.activities, *act)
		if len(r.activities) > r.cfg.ActivityCap {
			r.activities = r.activities[len(r.activities)-r.cfg.ActivityCap:]
		}
		payload := StatePayload{
			Room:         r.sanitizedInfoLocked(),
			State:        state,
			Activity:     act,
			Participants: cloneParticipants(r.info.Participants),
			Turn:         r.buildTurnStateLocked(),
		}
		env := envelopeWith(TypeSphUpdated, r.info.RoomID, "", r.version, payload)
		r.broadcastLocked(env)
		r.mu.Unlock()
		r.notifyChanged()
		return nil
	}
	if c != nil {
		return c.sendSyncPush(input)
	}
	return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
}

func (m *Manager) GetTurnState() *TurnState {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.room != nil {
		m.room.mu.Lock()
		defer m.room.mu.Unlock()
		return m.room.buildTurnStateLocked()
	}
	return m.clientTurn
}

// SendChatMessage mengirim pesan chat ke sesi aktif (host memproses lokal & broadcast,
// client meneruskan ke host). MessageType: text | system | master_data.
func (m *Manager) SendChatMessage(messageType, content, refPackage, refMeta string) error {
	messageType = strings.TrimSpace(messageType)
	if messageType == "" {
		messageType = ChatTypeText
	}
	m.mu.Lock()
	r, c := m.room, m.client
	m.mu.Unlock()

	cp := &ChatPayload{
		MessageID:   uuid.NewString(),
		MessageType: messageType,
		Content:     content,
		RefPackage:  refPackage,
		RefMeta:     refMeta,
	}

	switch {
	case r != nil:
		r.mu.Lock()
		host := r.hostParticipantLocked()
		r.mu.Unlock()
		if host == nil {
			return services.NewConflictError("Host room tidak ditemukan.")
		}
		r.handleIncomingChat(host, cp)
		return nil
	case c != nil:
		return c.sendChat(cp)
	default:
		return services.NewConflictError("Tidak ada sesi Work Together yang aktif.")
	}
}

// ClearChatUnread menetapkan unread chat menjadi 0 (saat chat dibuka).
func (m *Manager) ClearChatUnread() {
	m.mu.Lock()
	if m.client != nil {
		m.clientUnread = 0
		snap := m.sessionLocked()
		fn := m.emitFn
		m.mu.Unlock()
		if fn != nil {
			fn(snap)
		}
		return
	}
	if m.room != nil {
		m.hostUnread = 0
	}
	m.mu.Unlock()
}

// GetChatUnread mengembalikan jumlah pesan chat yang belum dibaca.
func (m *Manager) GetChatUnread() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		return m.clientUnread
	}
	if m.room != nil {
		return m.hostUnread
	}
	return 0
}
