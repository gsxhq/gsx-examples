package views

import (
	"bytes"
	"context"
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

func TestPatchTargetsByName(t *testing.T) {
	got := render(t, Patch("orders", gsx.Raw("<b>x</b>")))
	if !strings.Contains(got, `<template for="orders">`) {
		t.Errorf("missing template for=orders in %q", got)
	}
	if !strings.Contains(got, "<b>x</b>") {
		t.Errorf("patch body missing from %q", got)
	}
}

// A panel name is attacker-shaped input in the general case. `<template
// for={x}>` is an ordinary quoted-attribute sink, so gsx entity-escapes a `"`
// to `&#34;` rather than rejecting it — the name survives round-trip and
// still matches its region. This is a genuinely surprising asymmetry worth
// flagging: `<?start name={x}>` (the PI-name sink used by PanelShell) cannot
// escape a `"` inside processing-instruction data, so the same string hitting
// that sink is a render ERROR instead. Today every panel name is a static
// literal so nothing dynamic reaches either sink, but a future change that
// derives names from user input would land on one behavior or the other
// depending on which tag it flows through.
func TestPatchNameIsAttributeEscaped(t *testing.T) {
	got := render(t, Patch(`a"b`, gsx.Raw("x")))
	// Positive assertion: locks in that the quote is escaped in place, not
	// silently stripped. A regression that strips the quote (rendering
	// `for="ab"`) would corrupt the name — it would no longer match its
	// region, with no error to signal the mismatch — and only a positive
	// check on the exact escaped form catches that.
	if !strings.Contains(got, `for="a&#34;b"`) {
		t.Errorf("quote not escaped in place in %q", got)
	}
	if strings.Contains(got, `for="a"b"`) {
		t.Errorf("unescaped quote in %q", got)
	}
}

// A non-nil but empty Rows slice — what make([]Row, 0), a filter that empties
// a slice, or a JSON-decoded `[]` all produce — must still fall through to the
// value+badge shape, not render an empty, contentless table. This pins
// PanelBody's discriminator to len(p.Rows) > 0, not p.Rows != nil.
func TestPanelBodyEmptyRowsFallsThroughToValueShape(t *testing.T) {
	got := render(t, PanelBody(Panel{Rows: []Row{}, Value: "$0", Note: "flat"}))
	if strings.Contains(got, "<table") {
		t.Errorf("empty Rows rendered a table instead of the value+badge shape in %q", got)
	}
	if !strings.Contains(got, "$0") {
		t.Errorf("value+badge shape missing in %q", got)
	}
}

// PanelBody's waterfall bar renders BarPct as a "width:N%" style, via the
// explicit CSS attr literal — not silently rejected to gsx's ZgotmplZ CSS
// sanitizer placeholder, which is what a naive string-built style would hit.
func TestPanelBodyBarWidthReflectsBarPct(t *testing.T) {
	got := render(t, PanelBody(Panel{Value: "$0", Note: "flat", BarPct: 42}))
	if strings.Contains(got, "ZgotmplZ") {
		t.Fatalf("bar style was rejected by the CSS sanitizer in %q", got)
	}
	if !strings.Contains(got, `style="width:42%"`) {
		t.Errorf(`missing style="width:42%%" in %q`, got)
	}
}

func TestOrdinalSuffix(t *testing.T) {
	cases := map[int]string{
		1: "1st", 2: "2nd", 3: "3rd", 4: "4th",
		11: "11th", 12: "12th", 13: "13th",
		21: "21st", 22: "22nd", 23: "23rd", 24: "24th",
		101: "101st", 111: "111th", 112: "112th", 113: "113th",
	}
	for n, want := range cases {
		if got := ordinalSuffix(n); got != want {
			t.Errorf("ordinalSuffix(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestCircledNumeral(t *testing.T) {
	cases := map[int]string{1: "①", 2: "②", 3: "③"}
	for n, want := range cases {
		if got := circledNumeral(n); got != want {
			t.Errorf("circledNumeral(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestDocumentPositionMatchesPanelNamesOrder(t *testing.T) {
	for i, name := range PanelNames() {
		if got := documentPosition(name); got != i+1 {
			t.Errorf("documentPosition(%q) = %d, want %d", name, got, i+1)
		}
	}
}

func TestDocumentPositionPanicsOnUnknownName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("documentPosition did not panic on an unknown name")
		}
	}()
	documentPosition("no-such-panel")
}

// PanelShell's header now carries a second marker region, name+"-badge",
// holding a pending badge until the streaming server patches it — pin both
// the region and the pending copy, alongside the existing body region.
func TestPanelShellOpensAndClosesTheBadgeMarkerRegion(t *testing.T) {
	got := render(t, PanelShell("orders", "Recent orders"))
	if !strings.Contains(got, `<?start name="orders-badge">`) {
		t.Errorf("missing badge region open in:\n%s", got)
	}
	if !strings.Contains(got, "pending") {
		t.Errorf("missing pending copy in:\n%s", got)
	}
}

func TestPanelBadgeShowsOrdinalAndElapsed(t *testing.T) {
	got := render(t, PanelBadge(Panel{Ordinal: 1, ElapsedMS: 400}))
	if !strings.Contains(got, "1st · 400ms") {
		t.Errorf("missing ordinal/elapsed text in %q", got)
	}
}
