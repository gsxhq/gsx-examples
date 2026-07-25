package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A non-ResponseWriter must not error: the page has to render in tests and
// anywhere else that is not an HTTP response.
func TestFlushIsNoOpOffHTTP(t *testing.T) {
	var buf bytes.Buffer
	if err := Flush().Render(context.Background(), &buf); err != nil {
		t.Fatalf("Flush on a plain writer: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("Flush wrote %q, want nothing", buf.String())
	}
}

func TestFlushFlushesAResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := Flush().Render(context.Background(), rec); err != nil {
		t.Fatalf("Flush on a ResponseWriter: %v", err)
	}
	if !rec.Flushed {
		t.Fatal("ResponseWriter was not flushed")
	}
}

// unflushableWriter implements http.ResponseWriter but neither Flush nor
// FlushError, so http.NewResponseController.Flush reports ErrNotSupported.
// Pins the half of the guard that tolerates an unsupported writer.
type unflushableWriter struct{ header http.Header }

func (w *unflushableWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *unflushableWriter) Write(b []byte) (int, error) { return len(b), nil }

func (w *unflushableWriter) WriteHeader(int) {}

func TestFlushToleratesUnsupportedWriter(t *testing.T) {
	w := &unflushableWriter{}
	if err := Flush().Render(context.Background(), w); err != nil {
		t.Fatalf("Flush on an unflushable ResponseWriter: %v", err)
	}
}

// failingFlushWriter implements http.ResponseWriter and FlushError, always
// returning errFlush. Pins the half of the guard that propagates a genuine
// flush failure instead of swallowing it.
var errFlush = errors.New("flush: boom")

type failingFlushWriter struct{ header http.Header }

func (w *failingFlushWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingFlushWriter) Write(b []byte) (int, error) { return len(b), nil }

func (w *failingFlushWriter) WriteHeader(int) {}

func (w *failingFlushWriter) FlushError() error { return errFlush }

func TestFlushPropagatesAGenuineFlushError(t *testing.T) {
	w := &failingFlushWriter{}
	err := Flush().Render(context.Background(), w)
	if !errors.Is(err, errFlush) {
		t.Fatalf("Flush error = %v, want errors.Is(err, errFlush)", err)
	}
}
