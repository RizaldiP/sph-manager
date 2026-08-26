package collaboration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/RizaldiP/sph-manager/internal/services"
)

// clientParams adalah konfigurasi koneksi keluar ke room host.
type clientParams struct {
	addr        string
	displayName string
	deviceName  string
	accessCode  string
	roomCode    string

	cfg Config
	log *slog.Logger

	onStatus   func(status, errMsg string)
	onEnvelope func(env *Envelope)
	onClosed   func()
}

// connSession adalah satu koneksi WebSocket hidup beserta kanal kematianannya,
// sehingga read loop / heartbeat / supervisor bisa berhenti serempak tanpa double-close.
type connSession struct {
	conn     *websocket.Conn
	dead     chan struct{}
	deadOnce sync.Once
}

func (s *connSession) kill() {
	s.deadOnce.Do(func() {
		close(s.dead)
		_ = s.conn.Close()
	})
}

// Client adalah sesi aplikasi ini sebagai CLIENT pada room host:
// koneksi WebSocket keluar + auto-reconnect + heartbeat (§10.13–10.14).
type Client struct {
	p clientParams

	mu       sync.Mutex
	session  *connSession
	clientID string // identitas dari host; dipertahankan lintas reconnect
	status   string
	lastErr  string
	fatal    bool // true bila tidak perlu reconnect lagi (auth salah / room tutup)
	fatalErr string

	stopCh   chan struct{}
	stopOnce sync.Once

	firstResult chan error
	firstDone   bool
}

func newClient(p clientParams) (*Client, error) {
	if p.addr == "" {
		return nil, services.NewValidationError("Alamat host kosong.")
	}
	return &Client{
		p:           p,
		status:      ConnDisconnected,
		stopCh:      make(chan struct{}),
		firstResult: make(chan error, 1),
	}, nil
}

// StartAndWaitReady menjalankan supervisor dan menunggu initial sync pertama.
func (c *Client) StartAndWaitReady(timeout time.Duration) error {
	go c.run()
	select {
	case err := <-c.firstResult:
		if err != nil {
			c.p.log.Warn("join gagal (firstResult)", "addr", c.p.addr, "error", err)
		} else {
			c.p.log.Info("join berhasil (firstResult)", "addr", c.p.addr)
		}
		return err
	case <-time.After(timeout):
		c.p.log.Warn("join timeout", "addr", c.p.addr, "timeout", timeout)
		return fmt.Errorf("host di %s tidak merespons permintaan join (%s). "+
			"Kemungkinan firewall memblokir port. Izinkan SPH Manager pada jaringan privat.",
			c.p.addr, timeout)
	case <-c.stopCh:
		return fmt.Errorf("koneksi dibatalkan.")
	}
}

func (c *Client) resolveFirst(err error) {
	c.mu.Lock()
	already := c.firstDone
	c.firstDone = true
	c.mu.Unlock()
	if !already {
		c.firstResult <- err
	}
}

func (c *Client) setStatus(status, errMsg string) {
	c.mu.Lock()
	changed := c.status != status || errMsg != c.lastErr
	c.status = status
	if errMsg != "" {
		c.lastErr = errMsg
	}
	c.mu.Unlock()
	if changed && c.p.onStatus != nil {
		c.p.onStatus(status, errMsg)
	}
}

func (c *Client) currentSession() *connSession {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session
}

func (c *Client) setSession(s *connSession) {
	c.mu.Lock()
	c.session = s
	c.mu.Unlock()
}

func (c *Client) wsURL() string {
	return "ws://" + c.p.addr + "/ws"
}

// run adalah supervisor koneksi: dial → join → tahan koneksi → reconnect saat putus.
func (c *Client) run() {
	defer func() {
		if rec := recover(); rec != nil {
			c.p.log.Error("client.run panic", "recover", rec)
			c.setStatus(ConnDisconnected, "Terjadi kesalahan internal.")
			c.resolveFirst(fmt.Errorf("kesalahan internal client"))
		}
	}()
	cfg := c.p.cfg
	backoff := cfg.BackoffMin
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		c.mu.Lock()
		fatal := c.fatal
		c.mu.Unlock()
		if fatal {
			msg := c.fatalErr
			c.setStatus(ConnDisconnected, msg)
			c.resolveFirst(fmt.Errorf("%s", msg))
			return
		}

		if c.currentSession() == nil {
			s, err := c.connect()
			if err == nil {
				err = c.sendJoinOn(s)
			}
			if err == nil {
				c.setSession(s)
				go c.readLoop(s)
				go c.heartbeat(s)
				c.setStatus(ConnConnected, "")
				backoff = cfg.BackoffMin
			} else {
				s.kill()
				if isFatalClientErr(err) {
					c.mu.Lock()
					c.fatal = true
					c.fatalErr = err.Error()
					c.mu.Unlock()
					continue // ditangani di awal iterasi berikutnya
				}
				c.setStatus(ConnReconnecting, err.Error())
				c.p.log.Debug("gagal menyambung ke host", "error", err)
			}
		}

		cur := c.currentSession()
		if cur == nil {
			select {
			case <-c.stopCh:
				return
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > cfg.BackoffMax {
				backoff = cfg.BackoffMax
			}
			continue
		}

		// koneksi hidup — tunggu mati atau perintah berhenti
		select {
		case <-cur.dead:
			c.setStatus(ConnReconnecting, "")
		case <-c.stopCh:
			return
		}
		c.dropConn(cur)
		select {
		case <-c.stopCh:
			return
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > cfg.BackoffMax {
			backoff = cfg.BackoffMax
		}
	}
}

type clientErr struct {
	msg   string
	fatal bool
}

func (e *clientErr) Error() string { return e.msg }

func isFatalClientErr(err error) bool {
	ce, ok := err.(*clientErr)
	return ok && ce.fatal
}

func joinErrorMessage(ep ErrorPayload) string {
	if ep.Message != "" {
		return ep.Message
	}
	return "Host menolak permintaan join."
}

// connect membuka koneksi TCP/WebSocket baru.
func (c *Client) connect() (*connSession, error) {
	d := net.Dialer{Timeout: c.p.cfg.DialTimeout}
	dialer := websocket.Dialer{
		NetDialContext: d.DialContext,
		ReadBufferSize: 4096,
	}
	conn, _, err := dialer.DialContext(context.Background(), c.wsURL(), http.Header{})
	if err != nil {
		c.p.log.Warn("connect gagal", "addr", c.p.addr, "error", err)
		return nil, &clientErr{msg: fmt.Sprintf(
			"gagal terhubung ke %s: %s. Periksa: 1) IP dan port benar, "+
				"2) room host masih aktif, 3) firewall Windows mengizinkan aplikasi pada jaringan privat.",
			c.p.addr, err.Error())}
	}
	conn.SetReadLimit(maxMessageSize)
	return &connSession{conn: conn, dead: make(chan struct{})}, nil
}

// sendJoinOn mengirim JOIN_REQUEST pada sesi tertentu.
func (c *Client) sendJoinOn(s *connSession) error {
	jr := JoinRequest{
		ClientID:    c.clientIDLocked(),
		DisplayName: c.p.displayName,
		DeviceName:  c.p.deviceName,
		AccessCode:  c.p.accessCode,
		RoomCode:    c.p.roomCode,
	}
	return c.writeOn(s, envelopeWith(TypeJoinRequest, "", "", 0, jr))
}

func (c *Client) clientIDLocked() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clientID
}

// writeOn menulis satu envelope dengan penjagaan satu-penulis per sesi.
func (c *Client) writeOn(s *connSession, e *Envelope) error {
	if s == nil {
		return &clientErr{msg: "belum terhubung ke host."}
	}
	_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := s.conn.WriteJSON(e); err != nil {
		return &clientErr{msg: "pengiriman gagal: koneksi ke host terputus."}
	}
	return nil
}

// sendOp meneruskan operasi edit ke host.
func (c *Client) sendOp(op *services.OpPayload) error {
	if op == nil {
		return services.NewValidationError("Operasi kosong.")
	}
	c.mu.Lock()
	s := c.session
	c.mu.Unlock()
	if s == nil {
		return fmt.Errorf("operasi gagal dikirim: koneksi ke host sedang terputus.")
	}
	if err := c.writeOn(s, envelopeWith(TypeOpRequest, "", "", 0, op)); err != nil {
		return fmt.Errorf("operasi gagal dikirim: koneksi ke host sedang terputus.")
	}
	return nil
}

// readLoop menerima envelope dari host dan meneruskannya ke manager.
func (c *Client) readLoop(s *connSession) {
	defer func() {
		if rec := recover(); rec != nil {
			c.p.log.Error("client.readLoop panic", "recover", rec)
		}
	}()
	defer s.kill()
	for {
		_ = s.conn.SetReadDeadline(time.Now().Add(c.p.cfg.ReadWait))
		var e Envelope
		if err := s.conn.ReadJSON(&e); err != nil {
			return
		}

		switch e.Type {
		case TypeRoomJoined, TypeSyncResponse:
			if e.ClientID != "" {
				c.mu.Lock()
				c.clientID = e.ClientID
				c.mu.Unlock()
			}
			c.resolveFirst(nil)
		case TypeError:
			var ep ErrorPayload
			_ = json.Unmarshal(e.Payload, &ep)
			if ep.Code == errCodeAuth || ep.Code == errCodeNotJoined {
				c.markFatal(joinErrorMessage(ep))
			}
		case TypeRoomClosed:
			var cp ClosedPayload
			_ = json.Unmarshal(e.Payload, &cp)
			c.markFatal("Room telah ditutup oleh host." + tailNotice(cp.Reason))
		}

		if c.p.onEnvelope != nil {
			eCopy := e
			c.p.onEnvelope(&eCopy)
		}
	}
}

func tailNotice(reason string) string {
	if reason != "" {
		return " (" + reason + ")"
	}
	return ""
}

func (c *Client) markFatal(msg string) {
	c.mu.Lock()
	if !c.fatal {
		c.fatal = true
		c.fatalErr = msg
	}
	c.mu.Unlock()
}

// heartbeat mengirim PING periodik agar presence host tetap segar (§10.14).
func (c *Client) heartbeat(s *connSession) {
	defer func() {
		if rec := recover(); rec != nil {
			c.p.log.Error("client.heartbeat panic", "recover", rec)
		}
	}()
	ticker := time.NewTicker(c.p.cfg.Heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.writeOn(s, envelopeWith(TypePing, "", "", 0, nil)); err != nil {
				s.kill()
				return
			}
		case <-s.dead:
			return
		case <-c.stopCh:
			return
		}
	}
}

func (c *Client) dropConn(s *connSession) {
	c.mu.Lock()
	if c.session == s {
		c.session = nil
	}
	c.mu.Unlock()
	s.kill()
}

// stop memutus dan menghentikan seluruh loop; onClosed dipicu sekali.
func (c *Client) stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		c.mu.Lock()
		s := c.session
		c.mu.Unlock()
		if s != nil {
			// usahakan pamit sopan agar host langsung membersihkan presence
			_ = s.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			_ = s.conn.WriteJSON(envelopeWith(TypeLeave, "", "", 0, nil))
		}
		c.dropAll()
		if c.p.onClosed != nil {
			c.p.onClosed()
		}
	})
}

// stopQuiet seperti stop tetapi tanpa pamit maupun callback (kegagalan join awal).
func (c *Client) stopQuiet() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		c.dropAll()
	})
}

func (c *Client) dropAll() {
	c.mu.Lock()
	s := c.session
	c.session = nil
	c.mu.Unlock()
	if s != nil {
		s.kill()
	}
}
