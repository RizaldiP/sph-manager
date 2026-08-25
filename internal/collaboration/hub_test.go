package collaboration

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
