package main

import (
	"bytes"
	"context"
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
