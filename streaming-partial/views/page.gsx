package views

import (
	"github.com/gsxhq/gsx"
	"github.com/gsxhq/vite"
)

// PanelNames is the document order of the panels. Deliberately NOT the order
// they finish: the demo's whole point is that a later byte fills an earlier
// hole, so the slowest panel is first.
func PanelNames() []string { return []string{"revenue", "orders", "products"} }

// flush and stream are passed in rather than imported: both are server
// plumbing that lives in package main, and a view should not reach for the
// transport.
component Page(flush gsx.Node, stream gsx.Node) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>gsx — streaming partial updates</title>
			{{ v := vite.FromContext(ctx) }}
			{/* The scaffold's Layout (app.gsx) ships a dev-only FOUC gate here —
			    hide <body> via `html[data-loading]` until DOMContentLoaded, then
			    reveal, because Vite injects CSS via JS in dev. This page
			    deliberately OMITS it: DOMContentLoaded cannot fire until HTML
			    parsing completes, which for a streamed response means the
			    connection has closed — i.e. after all ~2.4s of panel latency, not
			    before. The gate would hide the entire progressive fill and reveal
			    everything at once, destroying the one thing this page exists to
			    demonstrate. The trade-off is a brief unstyled flash in dev before
			    Vite's injected CSS lands; that's a cosmetic blemish, not a defeat
			    of the demo, so it's the right one here. Do not restore this
			    block. */}
			{ if v.Dev() {
				{/* Same underlying incompatibility, second instance: in dev, Vite
				    ships CSS by injecting it from inside the "web/main.js" module
				    script below, and module scripts are deferred by spec — they
				    run only after HTML parsing completes, which for this
				    held-open stream means after the connection closes (~2.4s).
				    Left alone, that means no styles at all until the fill is
				    already done — measured live: the dev response's <head>
				    carries zero rel="stylesheet" links, only the module script.
				    The "?direct" suffix is Vite's dev-server switch that serves
				    the already-Tailwind-processed CSS as real text/css instead
				    of its normal text/javascript HMR-wrapped module, so a plain
				    <link> can load it in parallel with the stream like any other
				    stylesheet. Both files "web/main.js" imports are linked here,
				    in the same order, to match what v.Entry's real stylesheet
				    bundle contains in prod: style.css first (the scaffold's
				    base/reset — also sets body's flex-centering, which the
				    /scaffold page depends on), then gsxui.css (the
				    Tailwind-generated component styles this page's cards and
				    skeletons actually need). Linking only gsxui.css would still
				    style the cards correctly, but body's layout would jump the
				    instant main.js's deferred module finally ran and applied
				    style.css. Prod needs none of this: v.Entry already emits a
				    real <link rel="stylesheet">. */}
				<link rel="stylesheet" href="/__vite/web/style.css?direct"/>
				<link rel="stylesheet" href="/__vite/web/gsxui.css?direct"/>
			} }
			{{ assets := v.Entry("web/main.js") }}
			{ for _, href := range assets.CSS {
				<link rel="stylesheet" href={href}/>
			} }
			{ for _, src := range assets.Preloads {
				<link rel="modulepreload" href={src}/>
			} }
			{ for _, src := range assets.JS {
				<script type="module" src={src}></script>
			} }
			{/* The polyfill must run before any patch arrives, so it is not
			    deferred. It lives in the committed, embedded public/ dir, so it
			    needs no build step. */}
			<script src="/public/template-for-polyfill.js"></script>
		</head>
		{/* `block` beats web/style.css's `body { display: flex; place-items:
		    center }` (Vite's centred-splash default): a class selector
		    outranks an element selector regardless of source order, so this
		    one utility is what stops <main> being vertically centred as a
		    flex child. `min-h-screen` alone was never enough — it only ever
		    touched min-height, never display. */}
		<body class="block min-h-screen bg-background text-foreground p-8">
			<main class="mx-auto flex max-w-2xl flex-col gap-10">
				<div class="flex flex-col gap-2">
					<p class="font-mono text-xs text-muted-foreground">gsx · declarative partial updates</p>
					<h1 class="text-2xl font-semibold">A later byte fills an earlier hole.</h1>
				</div>
				<div class="flex flex-col gap-6">
					<PanelShell name="revenue" title="Revenue"/>
					<PanelShell name="orders" title="Recent orders"/>
					<PanelShell name="products" title="Top products"/>
				</div>
			</main>
			{/* Paint the skeletons before any panel work begins. */}
			{ flush }
			{ stream }
		</body>
	</html>
}
