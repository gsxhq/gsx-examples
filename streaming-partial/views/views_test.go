package views

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/gsxhq/gsx"
)

func render(t *testing.T, n gsx.Node) string {
	t.Helper()
	var buf bytes.Buffer
	if err := n.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

// noop stands in for the flush and stream nodes, which package main supplies.
func noop() gsx.Node {
	return gsx.Func(func(context.Context, io.Writer) error { return nil })
}

// PanelShell is asserted directly rather than through Page: Page reads the Vite
// asset bundle from context, so a full-document render needs a real *vite.Vite.
// The whole-page assertions live in Task 6's httptest, which has one.
func TestPanelShellOpensAndClosesAMarkerRegion(t *testing.T) {
	got := render(t, PanelShell("orders", "Recent orders"))
	if !strings.Contains(got, `<?start name="orders">`) {
		t.Errorf("missing region open in:\n%s", got)
	}
	if !strings.Contains(got, "<?end>") {
		t.Errorf("missing region close in:\n%s", got)
	}
	if !strings.Contains(got, "Recent orders") {
		t.Errorf("title missing in:\n%s", got)
	}
}

func TestPanelNamesAreTheDocumentOrder(t *testing.T) {
	want := []string{"revenue", "orders", "products"}
	got := PanelNames()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("PanelNames() = %v, want %v — the slowest panel must be first", got, want)
	}
}
