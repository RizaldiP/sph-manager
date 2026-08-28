package collaboration

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// ===== Unit tests: helpers.go =====

func TestGenerateRoomCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := GenerateRoomCode()
		if len(code) != 6 {
			t.Fatalf("GenerateRoomCode() len = %d, mau 6", len(code))
		}
		for _, c := range code {
			if !strings.ContainsRune(codeAlphabet, c) {
				t.Fatalf("GenerateRoomCode() mengandung karakter %q di luar alphabet", c)
			}
		}
		if seen[code] {
			t.Fatalf("GenerateRoomCode() duplikat: %q", code)
		}
		seen[code] = true
	}
}

func TestGenerateAccessCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := GenerateAccessCode()
		if len(code) != 6 {
			t.Fatalf("GenerateAccessCode() len = %d, mau 6", len(code))
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Fatalf("GenerateAccessCode() mengandung karakter non-digit %q", c)
			}
		}
		if seen[code] {
			t.Fatalf("GenerateAccessCode() duplikat: %q", code)
		}
		seen[code] = true
	}
}

func TestSanitizeIdentity(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"  hello  ", "hello"},
		{"hello   world", "hello world"},
		{"", ""},
		{"  ", ""},
		{"a b c", "a b c"},
		{strings.Repeat("x", 150), strings.Repeat("x", 100)},
	}
	for _, c := range cases {
		got := sanitizeIdentity(c.in)
		if got != c.want {
			t.Errorf("sanitizeIdentity(%q) = %q, mau %q", c.in, got, c.want)
		}
	}
}

func TestEqualConstTime(t *testing.T) {
	if !equalConstTime("123456", "123456") {
		t.Error("equalConstTime相同字符串 harus true")
	}
	if equalConstTime("123456", "123457") {
		t.Error("equalConstTime beda string harus false")
	}
}

func TestCloneParticipants(t *testing.T) {
	src := []Participant{
		{ID: "a", DisplayName: "Alice", Role: "HOST"},
		{ID: "b", DisplayName: "Bob", Role: "EDITOR"},
	}
	cloned := cloneParticipants(src)
	if len(cloned) != len(src) {
		t.Fatalf("cloneParticipants len = %d, mau %d", len(cloned), len(src))
	}
	cloned[0].DisplayName = "CHANGED"
	if src[0].DisplayName == "CHANGED" {
		t.Error("cloneParticipants harus buat independent copy")
	}
}

func TestCloneParticipantsNil(t *testing.T) {
	if cloned := cloneParticipants(nil); cloned != nil {
		t.Errorf("cloneParticipants(nil) = %v, mau nil", cloned)
	}
}

func TestCloneActivities(t *testing.T) {
	src := []services.CollabActivity{
		{Actor: "Alice", Action: "edit", Summary: "edited header"},
	}
	cloned := cloneActivities(src)
	cloned[0].Actor = "CHANGED"
	if src[0].Actor == "CHANGED" {
		t.Error("cloneActivities harus buat independent copy")
	}
}

func TestCloneRoomInfo(t *testing.T) {
	ri := &RoomInfo{
		RoomID:   "r1",
		RoomName: "Test Room",
		Participants: []Participant{
			{ID: "a", DisplayName: "Alice"},
		},
	}
	cloned := cloneRoomInfo(ri)
	cloned.RoomName = "CHANGED"
	if ri.RoomName == "CHANGED" {
		t.Error("cloneRoomInfo harus buat independent copy")
	}
	cloned.Participants[0].DisplayName = "CHANGED"
	if ri.Participants[0].DisplayName == "CHANGED" {
		t.Error("cloneRoomInfo.Participants harus independent copy")
	}
}

func TestCloneRoomInfoNil(t *testing.T) {
	if cloned := cloneRoomInfo(nil); cloned != nil {
		t.Errorf("cloneRoomInfo(nil) = %v, mau nil", cloned)
	}
}

func TestSortDiscoveredByNewest(t *testing.T) {
	now := time.Now()
	rows := []DiscoveredRoom{
		{RoomID: "old", LastSeen: now.Add(-2 * time.Hour)},
		{RoomID: "mid", LastSeen: now.Add(-1 * time.Hour)},
		{RoomID: "new", LastSeen: now},
	}
	sortDiscoveredByNewest(rows)
	if rows[0].RoomID != "new" || rows[1].RoomID != "mid" || rows[2].RoomID != "old" {
		t.Errorf("sortDiscoveredByNewest urutan salah: %v", rows)
	}
}

func TestStatusLabelID(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"DRAFT", "Draft"},
		{"REVIEW", "Review"},
		{"FINAL", "Final"},
		{"SENT", "Terkirim"},
		{"ACCEPTED", "Disetujui"},
		{"REJECTED", "Ditolak"},
		{"CANCELLED", "Dibatalkan"},
		{"UNKNOWN", "UNKNOWN"},
	}
	for _, c := range cases {
		got := statusLabelID(c.in)
		if got != c.want {
			t.Errorf("statusLabelID(%q) = %q, mau %q", c.in, got, c.want)
		}
	}
}

// ===== Unit tests: messages.go =====

func TestNewEnvelope(t *testing.T) {
	e := newEnvelope(TypeJoinRequest)
	if e.Type != TypeJoinRequest {
		t.Errorf("newEnvelope type = %q, mau %q", e.Type, TypeJoinRequest)
	}
	if e.MessageID == "" {
		t.Error("newEnvelope MessageID kosong")
	}
	if e.Timestamp.IsZero() {
		t.Error("newEnvelope Timestamp zero")
	}
}

func TestEnvelopeWith(t *testing.T) {
	payload := StatePayload{
		Room: &RoomInfo{RoomID: "r1", RoomName: "Test"},
	}
	e := envelopeWith(TypeSphUpdated, "r1", "c1", 42, payload)
	if e.RoomID != "r1" {
		t.Errorf("envelopeWith RoomID = %q, mau r1", e.RoomID)
	}
	if e.ClientID != "c1" {
		t.Errorf("envelopeWith ClientID = %q, mau c1", e.ClientID)
	}
	if e.Version != 42 {
		t.Errorf("envelopeWith Version = %d, mau 42", e.Version)
	}
	if e.Payload == nil {
		t.Fatal("envelopeWith Payload nil")
	}

	// unmarshal back
	var sp StatePayload
	if err := json.Unmarshal(e.Payload, &sp); err != nil {
		t.Fatalf("Payload unmarshal error: %v", err)
	}
	if sp.Room == nil || sp.Room.RoomID != "r1" {
		t.Errorf("Payload Room.RoomID = %v, mau r1", sp.Room)
	}
}

func TestEnvelopeWithNilPayload(t *testing.T) {
	e := envelopeWith(TypePong, "", "", 0, nil)
	if e.Payload != nil {
		t.Errorf("envelopeWith(nil payload) Payload = %v, mau nil", e.Payload)
	}
}

func TestEnvelopeMarshalRoundtrip(t *testing.T) {
	e := envelopeWith(TypeSphUpdated, "r1", "c1", 5, StatePayload{
		Room: &RoomInfo{RoomID: "r1"},
	})
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded Envelope
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.Type != TypeSphUpdated {
		t.Errorf("roundtrip type = %q, mau %q", decoded.Type, TypeSphUpdated)
	}
	if decoded.RoomID != "r1" {
		t.Errorf("roundtrip roomId = %q, mau r1", decoded.RoomID)
	}
}

func TestChatPayloadEnvelopeRoundtrip(t *testing.T) {
	cp := ChatPayload{
		MessageID:   "m1",
		RoomID:      "r1",
		SenderID:    "c1",
		SenderName:  "Budi",
		MessageType: ChatTypeMasterData,
		Content:     "Perbaikan Separator BBM",
		Status:      ChatStatusDelivered,
		RefPackage:  "pkg-1",
		CreatedAt:   time.Now(),
	}
	e := envelopeWith(TypeChatBroadcast, "r1", "", 0, cp)
	var decoded Envelope
	if err := json.Unmarshal(mustMarshal(t, e), &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	var out ChatPayload
	if err := json.Unmarshal(decoded.Payload, &out); err != nil {
		t.Fatalf("Payload unmarshal error: %v", err)
	}
	if out.MessageID != "m1" || out.MessageType != ChatTypeMasterData || out.RefPackage != "pkg-1" {
		t.Errorf("ChatPayload roundtrip salah: %+v", out)
	}
	if out.SenderName != "Budi" {
		t.Errorf("SenderName = %q, mau Budi", out.SenderName)
	}
}

func TestChatHistoryPayloadRoundtrip(t *testing.T) {
	hp := ChatHistoryPayload{Messages: []ChatPayload{
		{MessageID: "m1", MessageType: ChatTypeText, SenderName: "Admin"},
		{MessageID: "m2", MessageType: ChatTypeSystem, SenderName: ""},
	}}
	e := envelopeWith(TypeChatHistory, "r1", "", 0, hp)
	var decoded Envelope
	if err := json.Unmarshal(mustMarshal(t, e), &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	var out ChatHistoryPayload
	if err := json.Unmarshal(decoded.Payload, &out); err != nil {
		t.Fatalf("Payload unmarshal error: %v", err)
	}
	if len(out.Messages) != 2 || out.Messages[1].MessageID != "m2" {
		t.Errorf("ChatHistoryPayload roundtrip salah: %+v", out)
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	return b
}

// fakeChatStore adalah implementasi ChatStore in-memory untuk test.
type fakeChatStore struct {
	msgs []ChatMessage
}

func (f *fakeChatStore) SaveChat(m ChatMessage) error {
	f.msgs = append(f.msgs, m)
	return nil
}

func (f *fakeChatStore) History(roomID string, limit int) ([]ChatMessage, error) {
	var out []ChatMessage
	for _, m := range f.msgs {
		if m.RoomID == roomID {
			out = append(out, m)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

// TestHandleIncomingChatPersistenceDanLog memverifikasi host menyimpan chat ke
// store dan menambahkannya ke chatLog (untuk snapshot UI host).
func TestHandleIncomingChatPersistenceDanLog(t *testing.T) {
	store := &fakeChatStore{}
	mgr := NewManager(Config{}, nil, nil, nil)
	mgr.SetChatStore(store)

	// Simulasi room host tanpa server jaringan aktif.
	r := &Room{
		info:     RoomInfo{RoomID: "room-1", RoomName: "SPH KRI TBT", HostName: "Admin"},
		log:      testLogger(),
		mu:       sync.Mutex{},
		chat:     store,
		chatLog:  []ChatPayload{},
		chatCap:  100,
		byID:     map[string]*Participant{"host": {ID: "host", DisplayName: "Admin", Role: RoleHost}},
		conns:    map[string]*serverConn{},
		stopCh:   make(chan struct{}),
		closedCh: make(chan struct{}),
	}
	r.closeOnce = sync.Once{}

	host := r.hostParticipantLocked()
	cp := &ChatPayload{MessageID: "m-abc", MessageType: ChatTypeText, Content: "Halo semua"}
	r.handleIncomingChat(host, cp)

	if len(store.msgs) != 1 {
		t.Fatalf("chat tidak dipersist: store.msgs = %d", len(store.msgs))
	}
	if store.msgs[0].SenderName != "Admin" || store.msgs[0].MessageType != ChatTypeText {
		t.Errorf("isi store chat salah: %+v", store.msgs[0])
	}
	if len(r.chatLog) != 1 {
		t.Fatalf("chatLog host = %d, mau 1", len(r.chatLog))
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ===== Master Data (routing relay) =====

// fakeMasterDataStore adalah implementasi MasterDataStore in-memory untuk test.
type fakeMasterDataStore struct {
	Inbox   []models.MasterInbox
	Sent    []models.MasterSent
	inboxStatus map[string]string
	sentStatus  map[string]string
}

func (f *fakeMasterDataStore) SaveInbox(m *models.MasterInbox) error {
	for _, e := range f.Inbox {
		if e.PackageID == m.PackageID {
			return nil
		}
	}
	f.Inbox = append(f.Inbox, *m)
	return nil
}

func (f *fakeMasterDataStore) SetInboxStatus(packageID, status string, at time.Time) error {
	if f.inboxStatus == nil {
		f.inboxStatus = map[string]string{}
	}
	f.inboxStatus[packageID] = status
	return nil
}

func (f *fakeMasterDataStore) SaveSent(m *models.MasterSent) error {
	f.Sent = append(f.Sent, *m)
	return nil
}

func (f *fakeMasterDataStore) SetSentStatus(packageID, status string) error {
	if f.sentStatus == nil {
		f.sentStatus = map[string]string{}
	}
	f.sentStatus[packageID] = status
	return nil
}

func TestMasterDataTransferPayloadRoundtrip(t *testing.T) {
	tp := MasterDataTransferPayload{
		PackageID:     "md-1",
		SenderID:      "host",
		SenderName:    "Admin",
		SourceVersion: 3,
		Checksum:      "abc123",
		Title:         "Master Data (4 item)",
		Summary:       "Master Data (4 item)",
		Targets:       []string{"p1", "p2"},
		Payload:       json.RawMessage(`{"metadata":{},"data":{},"checksum":"abc123"}`),
	}
	e := envelopeWith(TypeMasterData, "room-1", "", 0, tp)
	var decoded Envelope
	if err := json.Unmarshal(mustMarshal(t, e), &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	var out MasterDataTransferPayload
	if err := json.Unmarshal(decoded.Payload, &out); err != nil {
		t.Fatalf("Payload unmarshal error: %v", err)
	}
	if out.PackageID != "md-1" || out.Checksum != "abc123" || out.SourceVersion != 3 {
		t.Errorf("roundtrip salah: %+v", out)
	}
	if len(out.Targets) != 2 || len(out.Payload) == 0 {
		t.Errorf("targets/payload salah: %+v", out)
	}
}

// TestHostSendMasterDataPersistSentDanHostInbox memverifikasi host yang mengirim
// Master Data mencatat sent dan, bila target = host, menyimpan ke inbox host.
func TestHostSendMasterDataPersistSentDanHostInbox(t *testing.T) {
	mdStore := &fakeMasterDataStore{}
	r := &Room{
		info:     RoomInfo{RoomID: "room-1", RoomName: "SPH KRI TBT", HostName: "Admin"},
		log:      testLogger(),
		mu:       sync.Mutex{},
		md:       mdStore,
		chatLog:  []ChatPayload{},
		chatCap:  100,
		byID:     map[string]*Participant{"host": {ID: "host", DisplayName: "Admin", Role: RoleHost}},
		conns:    map[string]*serverConn{},
		stopCh:   make(chan struct{}),
		closedCh: make(chan struct{}),
	}
	r.closeOnce = sync.Once{}

	pkg := &MasterDataPackage{
		Metadata: MasterPackageMetadata{
			PackageID: "md-9", SenderID: "host", SenderName: "Admin",
			RoomID: "room-1", SourceVersion: 2,
		},
		Data: MasterPackageData{Categories: []PackageCategory{{Code: "CAT-A", Name: "A"}}},
	}
	sum, _ := pkg.ComputeChecksum()
	pkg.Checksum = sum

	if err := r.hostSendMasterData(&MasterDataTransferPayload{
		PackageID:     pkg.Metadata.PackageID,
		SenderID:      pkg.Metadata.SenderID,
		SenderName:    pkg.Metadata.SenderName,
		SourceVersion: pkg.Metadata.SourceVersion,
		Checksum:      pkg.Checksum,
		Summary:       "Master Data (1 item)",
	}, []string{"host"}); err != nil {
		t.Fatalf("hostSendMasterData: %v", err)
	}

	if len(mdStore.Sent) != 1 {
		t.Fatalf("sent tidak tercatat: %d", len(mdStore.Sent))
	}
	if mdStore.Sent[0].PackageID != "md-9" || mdStore.Sent[0].Recipients == "" {
		t.Errorf("record sent salah: %+v", mdStore.Sent[0])
	}
	if len(mdStore.Inbox) != 1 || mdStore.Inbox[0].PackageID != "md-9" {
		t.Fatalf("host inbox tidak terisi: %+v", mdStore.Inbox)
	}
	// host juga menambah entri chat master_data di log.
	if len(r.chatLog) != 1 || r.chatLog[0].MessageType != ChatTypeMasterData {
		t.Errorf("chat master_data host tidak tercatat: %+v", r.chatLog)
	}
}

// TestSendMasterDataTanpaSesi memastikan pengiriman tanpa room/client aktif ditolak.
func TestSendMasterDataTanpaSesi(t *testing.T) {
	mgr := NewManager(Config{}, nil, nil, nil)
	pkg := &MasterDataPackage{
		Metadata: MasterPackageMetadata{PackageID: "md-0", SenderID: "host", SenderName: "Admin"},
	}
	if err := mgr.SendMasterData(pkg, []string{"host"}); err == nil {
		t.Fatal("tanpa sesi aktif harus ditolak")
	}
}

// TestOnClientEnvelopeTypeMasterDataSavesInbox memverifikasi client yang menerima
// paket menyimpan inbox via store dan menambah cache incoming.
func TestOnClientEnvelopeTypeMasterDataSavesInbox(t *testing.T) {
	mdStore := &fakeMasterDataStore{}
	mgr := NewManager(Config{}, nil, nil, nil)
	mgr.SetMasterDataStore(mdStore)
	mgr.mu.Lock()
	mgr.client = &Client{}
	mgr.cfg = mgr.cfg.withDefaults()
	mgr.mu.Unlock()

	env := envelopeWith(TypeMasterData, "room-1", "", 0, MasterDataTransferPayload{
		PackageID:  "md-5",
		SenderID:   "p1",
		SenderName: "Budi",
		Summary:    "Master Data (2 item)",
		Payload:    json.RawMessage(`{}`),
	})
	mgr.onClientEnvelope(env)

	if len(mdStore.Inbox) != 1 || mdStore.Inbox[0].PackageID != "md-5" {
		t.Fatalf("inbox client tidak tersimpan: %+v", mdStore.Inbox)
	}
	if len(mgr.clientIncoming) != 1 || mgr.clientIncoming[0].PackageID != "md-5" {
		t.Fatalf("clientIncoming tidak terisi: %+v", mgr.clientIncoming)
	}
	mgr.mu.Lock()
	incoming := len(mgr.clientIncoming)
	mgr.mu.Unlock()
	if incoming != 1 {
		t.Fatalf("incoming count = %d, mau 1", incoming)
	}
}
