package collaboration

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1 << 20 // 1 MB — payload operasi jauh di bawah ini
	sendBuffer     = 256
	writeWait      = 10 * time.Second
)

// wsServer adalah pembungkus HTTP+WebSocket untuk room host.
type wsServer struct {
	ln       net.Listener
	srv      *http.Server
	upgrader websocket.Upgrader

	onConn func(*websocket.Conn)
}

// startWSServer membuka listener pada port tertentu (0 = port acak, untuk test).
func startWSServer(port int, log *slog.Logger, onConn func(*websocket.Conn)) (*wsServer, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &wsServer{
		ln:     ln,
		onConn: onConn,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	s.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Warn("server kolaborasi berhenti", "error", err)
		}
	}()
	return s, nil
}

// Port mengembalikan port TCP aktual yang terikat (penting saat port=0).
func (s *wsServer) Port() int {
	if addr, ok := s.ln.Addr().(*net.TCPAddr); ok {
		return addr.Port
	}
	return 0
}

func (s *wsServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.srv.Shutdown(ctx)
	_ = s.ln.Close()
}

func (s *wsServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(maxMessageSize)
	s.onConn(conn)
}

// serverConn membungkus satu koneksi client remote dengan antrian kirim berurutan.
type serverConn struct {
	ws            *websocket.Conn
	send          chan *Envelope
	done          chan struct{}
	shutdownOnce  sync.Once
	participantID string
	closed        atomic.Int32
}

func newServerConn(ws *websocket.Conn) *serverConn {
	return &serverConn{
		ws:   ws,
		send: make(chan *Envelope, sendBuffer),
		done: make(chan struct{}),
	}
}

// deliver mengantrikan envelope tanpa memblokir; false berarti buffer penuh/sudah tutup.
// Aman dipanggil dari goroutine mana pun dan tidak pernah panic karena tidak pernah close(send).
func (c *serverConn) deliver(e *Envelope) bool {
	if c.closed.Load() == 1 {
		return false
	}
	select {
	case c.send <- e:
		return true
	case <-c.done:
		return false
	default:
		return false
	}
}

// writePump mengirim envelope satu per satu menjaga urutan broadcast.
func (c *serverConn) writePump(log *slog.Logger) {
	for {
		select {
		case e := <-c.send:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteJSON(e); err != nil {
				log.Debug("gagal menulis ke client", "error", err)
				c.shutdown()
				return
			}
		case <-c.done:
			return
		}
	}
}

// shutdown menutup koneksi sekali saja; writePump keluar lewat channel done.
func (c *serverConn) shutdown() {
	c.shutdownOnce.Do(func() {
		c.closed.Store(1)
		close(c.done)
		_ = c.ws.Close()
	})
}

// rejectUnjoined mengirim ERROR lalu menutup koneksi yang belum sah.
func rejectUnjoined(ws *websocket.Conn, ep ErrorPayload) {
	e := envelopeWith(TypeError, "", "", 0, ep)
	_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
	_ = ws.WriteJSON(e)
	_ = ws.Close()
}
