package main

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gsxhq/gsx-examples/streaming-partial/views"
)

func TestMaxLatencyReturnsTheSlowestPanel(t *testing.T) {
	s := &server{latency: map[string]time.Duration{
		"a": 10 * time.Millisecond,
		"b": 30 * time.Millisecond,
		"c": 20 * time.Millisecond,
	}}
	if got, want := s.maxLatency(), 30*time.Millisecond; got != want {
		t.Errorf("maxLatency() = %v, want %v", got, want)
	}
}

func TestBarPctIsProportionalAndClamped(t *testing.T) {
	hundred := 100 * time.Millisecond
	cases := []struct {
		name    string
		elapsed time.Duration
		max     time.Duration
		want    int
	}{
		{"zero elapsed", 0, hundred, 0},
		{"half", 50 * time.Millisecond, hundred, 50},
		{"exactly max", hundred, hundred, 100},
		{"jitter past max is clamped", 150 * time.Millisecond, hundred, 100},
		{"zero max guards div-by-zero", 50 * time.Millisecond, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := barPct(c.elapsed, c.max); got != c.want {
				t.Errorf("barPct(%v, %v) = %d, want %d", c.elapsed, c.max, got, c.want)
			}
		})
	}
}

// patchOrder returns the panel names for which a "-badge" template patch
// appears in got, in the order they appear — i.e. arrival order, since
// (*server).stream emits a panel's badge patch first, immediately on arrival.
func patchOrder(t *testing.T, got string, names []string) []string {
	t.Helper()
	type hit struct {
		name string
		pos  int
	}
	var hits []hit
	for _, name := range names {
		pos := strings.Index(got, `<template for="`+name+`-badge">`)
		if pos < 0 {
			t.Fatalf("missing badge patch for %q in:\n%s", name, got)
		}
		hits = append(hits, hit{name, pos})
	}
	for i := range hits {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].pos < hits[i].pos {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
	order := make([]string, len(hits))
	for i, h := range hits {
		order[i] = h.name
	}
	return order
}

// The demo's whole claim: arrival order is the INVERSE of document order.
// PanelNames() puts the slowest panel (revenue) first; this test uses tiny,
// scaled-down but still-inverted latencies (kept proportional to
// production's 2400/400/1200ms) so it runs in milliseconds instead of
// seconds while still exercising the real goroutine/channel/timing path in
// (*server).stream — not a mock of it.
func TestStreamAssignsOrdinalsInArrivalOrderNotDocumentOrder(t *testing.T) {
	s := &server{latency: map[string]time.Duration{
		"revenue":  24 * time.Millisecond,
		"orders":   4 * time.Millisecond,
		"products": 12 * time.Millisecond,
	}}
	var buf bytes.Buffer
	if err := s.stream(context.Background()).Render(context.Background(), &buf); err != nil {
		t.Fatalf("stream render: %v", err)
	}
	got := buf.String()

	gotOrder := patchOrder(t, got, views.PanelNames())
	wantOrder := []string{"orders", "products", "revenue"}
	if strings.Join(gotOrder, ",") != strings.Join(wantOrder, ",") {
		t.Fatalf("arrival order = %v, want %v (document order is %v — arrival must invert it)",
			gotOrder, wantOrder, views.PanelNames())
	}

	// The ordinal each badge reports must match this same arrival order,
	// independent of document position: orders is 1st, products is 2nd,
	// revenue is 3rd.
	for name, want := range map[string]string{"orders": "1st", "products": "2nd", "revenue": "3rd"} {
		badge := got[strings.Index(got, `<template for="`+name+`-badge">`):]
		if !strings.Contains(badge[:strings.Index(badge, "</template>")], want) {
			t.Errorf("%s badge missing ordinal %q in %q", name, want, badge)
		}
	}
}

// BarPct must reflect each panel's own elapsed time against the slowest
// panel's declared latency: revenue (the slowest) should land at ~100%,
// orders (six times faster) at a small fraction, products (half revenue's
// latency) at roughly half. Generous tolerances absorb real scheduling
// jitter without masking a genuinely wrong proportion.
func TestStreamBarPctIsProportionalToDeclaredMax(t *testing.T) {
	s := &server{latency: map[string]time.Duration{
		"revenue":  24 * time.Millisecond,
		"orders":   4 * time.Millisecond,
		"products": 12 * time.Millisecond,
	}}
	var buf bytes.Buffer
	if err := s.stream(context.Background()).Render(context.Background(), &buf); err != nil {
		t.Fatalf("stream render: %v", err)
	}
	got := buf.String()

	barPctOf := func(name string) int {
		t.Helper()
		start := strings.Index(got, `<template for="`+name+`">`)
		if start < 0 {
			t.Fatalf("missing body patch for %q", name)
		}
		section := got[start:]
		const marker = `style="width:`
		i := strings.Index(section, marker)
		if i < 0 {
			t.Fatalf("missing bar width for %q in %q", name, section)
		}
		section = section[i+len(marker):]
		pct := section[:strings.Index(section, "%")]
		n, err := strconv.Atoi(pct)
		if err != nil {
			t.Fatalf("parsing bar pct %q for %q: %v", pct, name, err)
		}
		return n
	}

	if pct := barPctOf("revenue"); pct < 90 {
		t.Errorf("revenue (slowest panel) BarPct = %d, want close to 100", pct)
	}
	if pct := barPctOf("orders"); pct < 5 || pct > 40 {
		t.Errorf("orders BarPct = %d, want a small fraction of revenue's", pct)
	}
	if pct := barPctOf("products"); pct < 30 || pct > 70 {
		t.Errorf("products BarPct = %d, want roughly half of revenue's", pct)
	}
}
