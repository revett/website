package site

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type hub struct {
	mu      sync.Mutex
	clients map[chan struct{}]bool
}

// Serve builds the site, serves dist/ on addr, and rebuilds + reloads the browser whenever anything
// in content/ or web/ changes.
func Serve(addr string) error {
	rebuild := func() error {
		if err := build(true); err != nil {
			return err
		}

		return runTailwind()
	}

	if err := rebuild(); err != nil {
		return err
	}

	h := &hub{clients: map[chan struct{}]bool{}}
	go watch(rebuild, h)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(distDir)))
	mux.HandleFunc("/__reload", h.handle)

	fmt.Printf("serving http://localhost%s\n", addr)

	return http.ListenAndServe(addr, mux)
}

func (h *hub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func fingerprint() string {
	hash := sha256.New()

	for _, dir := range []string{"content", "web"} {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			fmt.Fprintf(hash, "%s %d %d\n", path, info.Size(), info.ModTime().UnixNano())

			return nil
		})
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (h *hub) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ch := make(chan struct{}, 1)

	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}()

	for {
		select {
		case <-ch:
			fmt.Fprint(w, "data: reload\n\n")
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func runTailwind() error {
	bin := filepath.Join("bin", "tailwindcss")
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("bin/tailwindcss not found; run `make serve` so it gets downloaded")
	}

	cmd := exec.Command(bin, "-i", "web/css/input.css", "-o", filepath.Join(distDir, "style.css"))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running tailwind: %w\n%s", err, out)
	}

	return nil
}

func watch(rebuild func() error, h *hub) {
	last := fingerprint()

	for {
		time.Sleep(300 * time.Millisecond)

		next := fingerprint()
		if next == last {
			continue
		}
		last = next

		start := time.Now()
		if err := rebuild(); err != nil {
			fmt.Fprintln(os.Stderr, "rebuild failed:", err)
			continue
		}

		fmt.Printf("rebuilt in %s\n", time.Since(start).Round(time.Millisecond))
		h.broadcast()
	}
}
