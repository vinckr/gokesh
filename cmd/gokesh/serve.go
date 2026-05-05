package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/vinckr/gokesh/internal/build"
)

// reloadHub broadcasts SSE reload events to all connected browsers.
type reloadHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newReloadHub() *reloadHub {
	return &reloadHub{clients: make(map[chan struct{}]struct{})}
}

func (h *reloadHub) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *reloadHub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *reloadHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

const reloadScript = `<script>
(function(){
  var es = new EventSource('/__reload');
  es.onmessage = function(){ window.location.reload(); };
})();
</script>`

// injectingHandler wraps http.FileServer and injects the live-reload script
// into HTML responses. The script is NOT injected into built files on disk.
type injectingHandler struct {
	fs    http.Handler
	hub   *reloadHub
}

func (h *injectingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/__reload" {
		h.hub.sseHandler(w, r)
		return
	}
	rec := &responseRecorder{header: w.Header(), code: 200}
	h.fs.ServeHTTP(rec, r)

	ct := w.Header().Get("Content-Type")
	if strings.Contains(ct, "text/html") || strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" || !strings.Contains(r.URL.Path, ".") {
		body := string(rec.body)
		body = strings.Replace(body, "</body>", reloadScript+"\n</body>", 1)
		if !strings.Contains(body, "</body>") {
			body += reloadScript
		}
		w.WriteHeader(rec.code)
		fmt.Fprint(w, body)
		return
	}
	w.WriteHeader(rec.code)
	w.Write(rec.body)
}

func (h *reloadHub) sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.subscribe()
	defer h.unsubscribe(ch)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

type responseRecorder struct {
	header http.Header
	body   []byte
	code   int
}

func (r *responseRecorder) Header() http.Header        { return r.header }
func (r *responseRecorder) WriteHeader(code int)       { r.code = code }
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *responseRecorder) ReadFrom(src io.Reader) (int64, error) {
	b, err := io.ReadAll(src)
	r.body = append(r.body, b...)
	return int64(len(b)), err
}

func runServe(outDir, configPath, templatesDir, addr string) error {
	hub := newReloadHub()

	// Start the watcher in a goroutine; broadcast after each successful rebuild.
	go build.WatchWithCallback(outDir, configPath, templatesDir, func() {
		hub.broadcast()
	})

	handler := &injectingHandler{
		fs:  http.FileServer(http.Dir(outDir)),
		hub: hub,
	}

	slog.Info("serving with live reload", "addr", "http://localhost"+addr)
	return http.ListenAndServe(addr, handler)
}
