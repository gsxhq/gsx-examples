package main

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gsxhq/gsx"
)

// Flush streams whatever has been written so far to the client. It writes no
// markup of its own.
//
// http.NewResponseController walks the writer's Unwrap() chain and prefers a
// FlushError() error over the legacy http.Flusher, so a buffering middleware is
// transparent. A writer that cannot flush reports ErrNotSupported, which is a
// best-effort no-op — that is what lets the page render into a bytes.Buffer in
// tests.
func Flush() gsx.Node {
	return gsx.Func(func(ctx context.Context, w io.Writer) error {
		rw, ok := w.(http.ResponseWriter)
		if !ok {
			return nil
		}
		if err := http.NewResponseController(rw).Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
			return err
		}
		return nil
	})
}
