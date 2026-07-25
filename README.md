# gsx-examples

Runnable example projects for [gsx](https://github.com/gsxhq/gsx) — a JSX-like
templating language for Go that compiles to plain Go and streams HTML.

Each example is its own Go module, so one example's dependencies never constrain
another's. Clone the repo, `cd` into an example, and follow its README.

## Examples

### [streaming-partial](streaming-partial/) — streaming flush + declarative partial updates

A page streams its layout, flushes it, then fills the panels **out of document
order**: the slowest panel sits first, so a later byte fills an earlier hole.

![The three panels filling out of order: the middle panel arrives first, while the panel above it is still a skeleton](streaming-partial/docs/demo.gif)

Combines gsx's processing instructions (`<?marker>`, `<?start>…<?end>`) with the
streaming-flush pattern and [gsxui](https://github.com/gsxhq/gsxui) components.
Works natively in Chrome 148 behind the experimental-web-platform-features flag,
and everywhere else via a vendored polyfill.
