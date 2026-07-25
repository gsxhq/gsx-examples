package views

import (
	"strconv"

	"github.com/gsxhq/gsx-examples/streaming-partial/ui"
)

// PanelShell is a panel before its data arrives: a card whose header shows
// the panel's fixed document-position numeral plus a pending arrival badge,
// and whose body is a marker region holding a skeleton. Two
// `<template for="…">`s replace the two regions when the data lands: one for
// the badge (name+"-badge"), one for the body (name).
component PanelShell(name string, title string) {
	{{ pos := documentPosition(name) }}
	<ui.Card>
		<ui.CardHeader class="flex flex-row items-center justify-between gap-2">
			<div class="flex items-baseline gap-2">
				<span class="font-mono text-muted-foreground">{ circledNumeral(pos) }</span>
				<ui.CardTitle>{ title }</ui.CardTitle>
			</div>
			<?start name={name + "-badge"}>
				<PanelBadgePending/>
			<?end>
		</ui.CardHeader>
		<ui.CardContent>
			<?start name={name}>
				<ui.Skeleton class="h-24 w-full"/>
			<?end>
		</ui.CardContent>
	</ui.Card>
}

// Panel is one panel's rendered data plus the delivery measurements the
// streaming server records as it arrives. Ordinal, ElapsedMS, and BarPct are
// all zero until the panel has actually landed — panelData never sets them,
// only (*server).stream does, once it knows them.
type Panel struct {
	Name  string
	Title string
	Value string
	Note  string
	Rows  []Row

	// Ordinal is this panel's 1-based arrival order (1st, 2nd, 3rd…) — never
	// the same as its document position, that disagreement is the demo.
	Ordinal int
	// ElapsedMS is the measured wall-clock time, in milliseconds, from
	// request start to when this panel's patch was emitted.
	ElapsedMS int
	// BarPct is ElapsedMS as a percentage of the slowest panel's declared
	// latency, clamped to [0,100] — the waterfall bar's width.
	BarPct int
}

// Row is one line in a panel that shows a table.
type Row struct {
	Label string
	Value string
}

// PanelBadgePending is a panel's arrival badge before its data lands: the
// same header slot as PanelBadge, muted rather than accented since there is
// nothing to report yet.
component PanelBadgePending() {
	<ui.Badge variant="outline" class="font-mono text-muted-foreground">pending</ui.Badge>
}

// PanelBadge is a panel's arrival badge once its data has landed: its
// arrival ordinal and measured latency, both in the mono measurement layer,
// both in the one accent color the page spends on delivery timing. This
// number is what can disagree with PanelShell's document-position numeral —
// that disagreement is the entire point of the page.
component PanelBadge(p Panel) {
	<ui.Badge variant="outline" class="font-mono text-[oklch(0.72_0.13_195)] border-[oklch(0.72_0.13_195)]">
		{ ordinalSuffix(p.Ordinal) } · { strconv.Itoa(p.ElapsedMS) }ms
	</ui.Badge>
}

component PanelBody(p Panel) {
	// len, not nil-ness: a non-nil but empty Rows (make([]Row, 0), a filter
	// that empties a slice, a JSON-decoded `[]`) must still fall through to
	// the value+badge shape below, not render an empty, contentless table.
	{ if len(p.Rows) > 0 {
		<ui.Table>
			<ui.TableBody>
				{ for _, r := range p.Rows {
					<ui.TableRow>
						<ui.TableCell>{ r.Label }</ui.TableCell>
						<ui.TableCell class="text-right">{ r.Value }</ui.TableCell>
					</ui.TableRow>
				} }
			</ui.TableBody>
		</ui.Table>
	} else {
		<div class="flex items-baseline gap-3">
			<span class="text-3xl font-semibold">{ p.Value }</span>
			<ui.Badge variant="secondary">{ p.Note }</ui.Badge>
		</div>
	} }
	{/* The waterfall bar: width is this panel's own measured elapsed time as
	    a percentage of the slowest panel's declared latency (BarPct, set by
	    (*server).stream). Rendered via an explicit CSS attr literal so the
	    numeric hole is formatted and CSS-filtered by gsx rather than passed
	    through as an opaque interpolated string — a plain
	    `style={fmt.Sprintf("width:%d%%", p.BarPct)}` string would hit gsx's
	    strict CSS-value sanitizer as an unrecognized token and render
	    "ZgotmplZ" instead of a width. */}
	<div class="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
		<div class="h-full rounded-full bg-[oklch(0.72_0.13_195)]" style=css`width:@{p.BarPct}%`></div>
	</div>
}

// ordinalSuffix formats n with its English ordinal suffix: 1st, 2nd, 3rd,
// 4th, … and the 11th-13th exception (11th, not 11st; 12th, not 12nd; 13th,
// not 13rd).
func ordinalSuffix(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return strconv.Itoa(n) + "th"
	}
	switch n % 10 {
	case 1:
		return strconv.Itoa(n) + "st"
	case 2:
		return strconv.Itoa(n) + "nd"
	case 3:
		return strconv.Itoa(n) + "rd"
	default:
		return strconv.Itoa(n) + "th"
	}
}

// circledNumeral renders n as a Unicode circled digit (① ② ③ …), marking a
// panel's fixed position in document order — the number that never moves,
// unlike PanelBadge's Ordinal. Unicode assigns circled digits 1-20 as a
// contiguous block starting at U+2460; PanelNames() today has 3 entries,
// comfortably inside it.
func circledNumeral(n int) string {
	return string(rune('①' + n - 1))
}

// documentPosition returns name's 1-based index in PanelNames(), the single
// source of document order. name is always one of the small set of internal
// literals PanelShell's call sites pass, never user input, so an unknown
// name is a programmer error, not a runtime condition to recover from.
func documentPosition(name string) int {
	for i, n := range PanelNames() {
		if n == name {
			return i + 1
		}
	}
	panic("views: unknown panel name " + strconv.Quote(name))
}
