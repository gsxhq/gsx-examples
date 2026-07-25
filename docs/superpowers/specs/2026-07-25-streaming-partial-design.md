# streaming-partial: streaming flush + declarative partial updates

Status: approved (design), 2026-07-25

## Summary

A runnable demo combining three things gsx shipped recently:

- **processing instructions** — `<?marker name=…>` and `<?start name=…> … <?end>`
- **the streaming-flush pattern** — a `<Flush/>` custom node that pushes bytes
  to the client mid-render
- **gsxui** — shadcn-style components, vendored the way a real user vendors them

It is the first example in the `gsx-examples` sibling repo.

## What it proves

Plain HTTP streaming can only **append in document order**. Processing-instruction
markers let a *later* byte fill an *earlier* hole, because a `<template for="x">`
patches the marker named `x` wherever it already sits in the document.

So the demo inverts latency deliberately — the slowest panel is at the **top**:

| panel          | position | latency |
| -------------- | -------- | ------- |
| Revenue        | top      | 2.4 s   |
| Recent orders  | middle   | 0.4 s   |
| Top products   | bottom   | 1.2 s   |

Orders fills first while the space *above* it is still a skeleton, then products,
then revenue. That out-of-order fill is impossible with `<Flush/>` alone, and it
is the entire point of the demo.

## Why a sibling repo, not `gsx/examples/`

Two reasons, both discovered while designing:

1. It depends on **gsxui**, which depends on gsx. Putting it inside gsx would
   make the gsx repo carry a module pointing at its own downstream UI library.
2. gsx's `examples/` are pure Go with no node toolchain (`examples/tailwind-merge`
   is the precedent, and gsx CI lints it). gsxui needs a Tailwind v4 build, which
   would impose node on gsx's CI.

A standalone repo avoids both and can grow more examples later.

## Layout

Bootstrapped with `gsx init` (the blessed app scaffold: `@gsxhq/vite-plugin-gsx`,
`gsx dev` with HMR, and a `main.go` built on `github.com/gsxhq/vite` that embeds
`dist/` and serves hashed assets through `v.StaticHandler()`), then layered with
`gsxui init` + `gsxui add`.

```
gsx-examples/
  README.md                       what lives here
  streaming-partial/
    go.mod                        gsx >= v0.0.0-20260725075407-d72162ac2ac1, gsxui, gsxhq/vite
    main.go                       from gsx init; extended with the streaming handler
    flush.go                      the <Flush/> node
    data.go                       deterministic panel data
    vite.config.ts                from gsx init; + @tailwindcss/vite
    package.json                  from gsx init; + tailwindcss, tw-animate-css
    gsx.toml  gsxui.json          from gsxui init
    ui/                           vendored by `gsxui add card skeleton table badge`
    web/
      main.js                     entry; imports the gsxui css + js
      gsxui.css                   from gsxui init (Tailwind source)
    views/
      page.gsx                    document shell + three marker regions
      panels.gsx                  panel shell and real content
      patch.gsx                   <template for={name}> wrapper
    public/
      template-for-polyfill.js    vendored, 2.5 KB, Apache-2.0 (+ NOTICE)
    dist/                         COMMITTED build output (see below)
    README.md                     `go run .`, `npm run dev`, the Chrome flag
```

Each example is its own Go module, so one example's dependencies never constrain
another's.

`public/` is committed and embedded (`//go:embed all:public`), which is why the
polyfill lives there — it needs no build step at all.

## Render flow

1. Render the shell: `<head>` (CSS + polyfill), header, and three regions, each
   `<?start name="…"> <Skeleton/> <?end>`.
2. `<Flush/>` — the browser paints the full skeleton layout immediately.
3. Three goroutines produce panel data with staggered sleeps, delivered over a
   channel as they complete.
4. For each result, render `<template for="…">` with the real panel content,
   then `<Flush/>` again.
5. Close the document.

`gsx.Writer` is unbuffered and `Writer.Node` renders children against the
underlying `io.Writer`, so the flush boundary is exactly the bytes written so
far. `Flush()` uses `http.NewResponseController(rw).Flush()`, which walks the
`Unwrap()` chain and prefers `FlushError() error` over the legacy
`http.Flusher`, so a buffering middleware is transparent.

## Browser support

Declarative partial updates are **Chrome 148 behind
`chrome://flags/#enable-experimental-web-platform-features`**. Without support,
`<?marker …>` is an invisible comment and `<template for>` is inert — the page
would sit on its skeletons forever.

The demo therefore ships
[`template-for-polyfill`](https://github.com/GoogleChromeLabs/template-for-polyfill)
(2.5 KB, Apache-2.0), vendored as a single committed file under `public/`, so it
needs no build step of its own. Native path when the flag is on, polyfill
otherwise.

**Why the polyfill works at all** — non-obvious and worth recording: it walks
`SHOW_COMMENT | SHOW_PROCESSING_INSTRUCTION`. A browser without native PI
parsing turns `<?marker name="x">` into a *bogus comment* whose data is
`?marker name="x"`, so the polyfill reconstructs the target from either node
type. It also depth-counts `?start` for regions. The same server bytes work on
both paths — the server never negotiates.

## gsxui usage

Vendored the real way: `gsxui init` (writes `gsxui.css`, `gsxui.json`, and
`gsx.toml` wiring) then `gsxui add card skeleton table badge`. The vendored
`.gsx` sources sit in `ui/` and are readable in-tree, so the example teaches the
actual copy-in workflow rather than an import shortcut.

**No gsxui bump is required.** gsx is untagged, so versions are pseudo-versions.
gsxui pins `v0.0.0-20260724160502-9bb55ae38eec` (pre-PI); the demo requires
`v0.0.0-20260725075407-d72162ac2ac1` (the processing-instruction merge) directly,
and Go's minimal-version-selection picks the higher one, so gsxui compiles
against it. Processing instructions are purely
additive. If that assumption is wrong the build fails immediately and loudly —
which is the right failure mode.

## CSS

gsxui components carry Tailwind v4 utility classes and `gsxui.css` is
`@import "tailwindcss"`, so styling requires a build. It runs through Vite
(`@tailwindcss/vite`) — the same way gsxui's own site does it — rather than a
one-off CLI invocation.

`gsx init` gitignores `/dist/*`, which would mean a fresh clone serves an
unstyled page until someone runs npm. Since the demo's value is that a reader
can watch it, **`dist/` is deliberately un-ignored and committed**. `go run .`
then works straight from a clone; `npm run dev` remains available for the full
HMR loop.

**Known risks, both accepted:**

- Committed build output can go stale if `web/` or the vendored components
  change. Mitigated by a documented rebuild command, not a guarantee.
- `*.x.go` stays gitignored, per the scaffold's convention for gsx *apps* — the
  Vite plugin generates on the fly. So a clone needs `gsx generate` (or
  `npm run dev`) before `go run .` compiles. The README must say so plainly;
  this is the one thing a Go-only reader still has to run.

## Testing

The claim under test is an *ordering* claim, so the test asserts byte order, not
just content. Using `httptest`:

- the shell and all three skeleton regions appear before any `<template for=…>`;
- the three `<template>` blocks appear in **latency** order (orders, products,
  revenue), not document order;
- every marker name emitted by the shell is patched exactly once.

Latencies are injected (a field on the server struct, not a package constant) so
the suite runs in milliseconds rather than sleeping 2.4 s.

A second test renders through a writer that is **not** an `http.ResponseWriter`
and asserts the page still renders correctly — `Flush()` must degrade to a no-op.

## Non-goals

- Hosting: no Dockerfile, no fly.toml. A live URL would need the edge proxy
  verified not to buffer, which is its own piece of work.
- No database; panel data is generated.
- No `html-setters-polyfill` — the demo does not use those APIs.
- No CI beyond the repo's own `go test` and lint.
