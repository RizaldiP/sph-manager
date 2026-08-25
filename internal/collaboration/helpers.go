package collaboration

import (
	"crypto/rand"
	"crypto/subtle"
	"math/big"
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

// sortDiscoveredByNewest mengurutkan entri discovery dari yang terakhir terlihat.
func sortDiscoveredByNewest(rows []DiscoveredRoom) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].LastSeen.After(rows[j-1].LastSeen); j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
