package main

import (
	"cmp"
	"context"
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gsxhq/gsx"
	"github.com/gsxhq/gsx-examples/streaming-partial/views"
	"github.com/gsxhq/vite"
)

//go:embed all:dist
var distFS embed.FS

//go:embed all:public
var publicFS embed.FS

type server struct{ latency map[string]time.Duration }

func newServer() *server {
	// Deliberately inverted against document order: the FIRST panel is the
	// slowest, so the page visibly fills out of order.
	return &server{latency: map[string]time.Duration{
		"revenue":  2400 * time.Millisecond,
		"orders":   400 * time.Millisecond,
		"products": 1200 * time.Millisecond,
	}}
}

// maxLatency returns the slowest panel's declared latency — the denominator
// for each panel's waterfall-bar BarPct. It is the map's own known ceiling
// (2400ms today, revenue's), not a value tracked from measured arrivals.
func (s *server) maxLatency() time.Duration {
	var max time.Duration
	for _, d := range s.latency {
		if d > max {
			max = d
		}
	}
	return max
}

// barPct returns elapsed as a percentage of max, clamped to [0,100]. A panel
// arriving right on its declared latency reads at or near 100%; the clamp
// only guards against the measured elapsed (real wall-clock time, including
// goroutine-scheduling jitter) overshooting the declared max by a few
// milliseconds and pushing the bar past its own track.
func barPct(elapsed, max time.Duration) int {
	if max <= 0 {
		return 0
	}
	pct := int(elapsed * 100 / max)
	if pct > 100 {
		pct = 100
	}
	return pct
}

// stream emits, per panel in completion order: a <template for="…-badge">
// carrying its arrival ordinal and measured latency, then a
// <template for="…"> carrying its body content and waterfall bar — flushing
// after each pair so the browser applies it immediately. It blocks until
// every panel has been sent, which is what keeps the response open.
func (s *server) stream(ctx context.Context) gsx.Node {
	return gsx.Func(func(ctx context.Context, w io.Writer) error {
		start := time.Now()
		max := s.maxLatency()
		results := make(chan views.Panel, len(views.PanelNames()))
		for _, name := range views.PanelNames() {
			go func(name string) {
				select {
				case <-time.After(s.latency[name]):
					results <- panelData(name)
				case <-ctx.Done():
				}
			}(name)
		}
		ordinal := 0
		for range views.PanelNames() {
			select {
			case p := <-results:
				ordinal++
				elapsed := time.Since(start)
				p.Ordinal = ordinal
				p.ElapsedMS = int(elapsed.Milliseconds())
				p.BarPct = barPct(elapsed, max)

				badge := views.Patch(p.Name+"-badge", views.PanelBadge(p))
				if err := badge.Render(ctx, w); err != nil {
					return err
				}
				body := views.Patch(p.Name, views.PanelBody(p))
				if err := body.Render(ctx, w); err != nil {
					return err
				}
				if err := Flush().Render(ctx, w); err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Proxies that buffer would defeat the whole demo; ask them not to.
	w.Header().Set("X-Accel-Buffering", "no")
	page := views.Page(Flush(), s.stream(r.Context()))
	if err := page.Render(r.Context(), w); err != nil {
		// The header is already sent by now, so an error cannot become a 500 —
		// log it and let the truncated response speak for itself.
		log.Printf("render: %v", err)
	}
}

func main() {
	devURL := os.Getenv("VITE_DEV_URL") // "" in prod
	v, err := vite.New(vite.Config{DevURL: devURL, DevBase: "/__vite/", Dist: distFS, DistDir: "dist"})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/public/", http.FileServerFS(publicFS))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if !v.Dev() {
		mux.Handle("/static/", v.StaticHandler())
	}
	s := newServer()
	// The demo owns "/" — it is what this example exists to show.
	mux.HandleFunc("/", s.handle)
	// The scaffold's landing page, kept reachable so its styling stays live.
	mux.HandleFunc("/scaffold", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := Index("gsx + Vite").Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// v.Middleware injects *vite.Vite into each request's context so components
	// read the asset bundle from ctx (no prop threading).
	port := cmp.Or(os.Getenv("GO_PORT"), "7777")
	srv := &http.Server{Addr: ":" + port, Handler: v.Middleware(mux)}

	// Serve in the background so the main goroutine can wait for a shutdown
	// signal. gsx dev sends SIGTERM on each rebuild; shutting down gracefully
	// releases the port BEFORE exit, so the next build re-binds cleanly.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Printf("listening on http://localhost:%s", port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
