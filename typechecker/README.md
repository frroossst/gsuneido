# SuneidoTypes

A type inference engine and checker for [Suneido](https://suneido.com) code.
Suneido is dynamically typed; this tool infers types for params, locals, and
class members by walking the gsuneido AST, then reports diagnostics (bad
operands, missing members, arity mismatches, etc.) with a confidence score so
an IDE can filter out the guesswork.

## Usage

The binary reads a JSON request on stdin and writes diagnostics to stdout.
`method` is `TypeInfer` (report diagnostics) or `TypeAnnotate` (splice
inferred annotations back into the source):

    echo '{"method": "TypeInfer", "arguments": ["class { F() { return 1 + 2 } }"]}' | bin/suneidotypes

Pass `-serve` to run it as a long-lived HTTP server instead (prints
`READY port=N` on startup). For local dev, `runner.py` wraps the JSON
protocol: point it at a `.ss` / `.suneido` file and it pretty-prints the
findings (`--annotate` for the annotated source).

## Building and testing

    make build     # builds bin/suneidotypes (plus a linux binary)
    make test      # all tests, including one round of property-based tests
    make pbt       # just the property-based tests (rapid)
    make slowtest  # property tests with many more iterations

See the Makefile for deploy and the continuous property-test loop.

## Layout

- `typeinfer/` - public entry point (`Process`), session handling
- `typeinfer/internal/engine/` - the passes: local inference, flow narrowing,
  requirement inference, callsite/capability/arity checks
- `typeinfer/typealgebra/` - the type lattice: primitives, unions, dirty types
- `typeinfer/annotations/` - builtin signature database plus a hand-written overlay
- `typeinfer/diagnostics/` - severity, confidence scoring, filtering
- `cmd/suneidotypes/` - the binary
