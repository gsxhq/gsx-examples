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
			{ if v.Dev() {
				<style>
					html[data-loading] body {
						visibility: hidden;
					}

					html[data-loading] * {
						transition: none !important;
					}
				</style>
				<script>
					// Dev-only FOUC gate. Vite injects CSS via JS after the HTML
					// loads, so hide the page until every module script has run
					// (DOMContentLoaded) and one paint has landed (double rAF),
					// then reveal. Prod ships real <link rel=stylesheet> tags
					// below, so no gate is emitted there.
					document.documentElement.dataset.loading = "true";
					var unhide = function () {
						document.documentElement.removeAttribute("data-loading");
					};
					var reveal = function () {
						requestAnimationFrame(function () { requestAnimationFrame(unhide); });
					};
					if (document.readyState === "loading") {
						document.addEventListener("DOMContentLoaded", reveal);
					} else {
						reveal();
					}
					// Safety net (rAF pauses in background tabs).
					setTimeout(unhide, 5000);
				</script>
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
		<body class="min-h-screen bg-background text-foreground p-8">
			<h1 class="text-2xl font-semibold mb-6">Streaming partial updates</h1>
			<main class="grid gap-6 max-w-3xl">
				<PanelShell name="revenue" title="Revenue"/>
				<PanelShell name="orders" title="Recent orders"/>
				<PanelShell name="products" title="Top products"/>
			</main>
			{/* Paint the skeletons before any panel work begins. */}
			{ flush }
			{ stream }
		</body>
	</html>
}
