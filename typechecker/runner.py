#!/usr/bin/env python
# NOTE: this script was LLM generated
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path


CONFIG = {
    "strictStringConcat":      "off",
    "strictCrossTypeCompares": "off",
}


ANSI_RESET = "\033[0m"
ANSI_BOLD = "\033[1m"
ANSI_RED = "\033[31m"
ANSI_YELLOW = "\033[33m"
ANSI_GREEN = "\033[32m"


def colored(code: str, s: str, enabled: bool) -> str:
    return f"{code}{s}{ANSI_RESET}" if enabled else s


def find_binary() -> str:
    """Locate the compiled suneidotypes binary next to this script.

    Tries the Windows .exe first (since this repo is developed on
    Windows / Git Bash), then the bare name. Falls back to PATH lookup
    as a last resort so a `make deploy`d binary still works.

    Before returning, ensures the binary is fresh w.r.t. the local Go
    sources (rebuilds if any non-test .go file is newer). Stale binaries
    were too easy to hit: edit a typeinfer pass, rerun, debug a panic that
    was already fixed in the working tree. The freshness check is a
    no-op when nothing changed.
    """
    here = Path(__file__).resolve().parent
    bin_dir = here / "bin"
    local = bin_dir / ("suneidotypes.exe" if os.name == "nt" else "suneidotypes")
    if local.exists() or (here / "cmd" / "suneidotypes").is_dir():
        ensure_fresh_binary(here, local)
        if local.exists():
            return str(local)
    # Fall back to a deployed binary on PATH. We can't freshness-check that
    # one (no source tree co-located), so accept it as-is.
    from shutil import which
    found = which("suneidotypes") or which("suneidotypes.exe")
    if found:
        return found
    sys.exit("error: could not find bin/suneidotypes - run `make build` first")


def ensure_fresh_binary(repo_root: Path, binary: Path) -> None:
    """Rebuild `binary` if any non-test .go file under repo_root is newer.

    Why this exists: the binary lives in bin/, but every dev workflow
    edits .go files under cmd/ and typeinfer/. Without a freshness check,
    `python runner.py X` happily runs against a months-old build and you
    waste time chasing a panic that's already fixed at HEAD. Auto-rebuild
    keeps the wrapper script honest with zero ceremony.

    Test files are skipped: they never link into the production binary,
    so editing one shouldn't trigger a rebuild.
    """
    if binary.exists():
        bin_mtime = binary.stat().st_mtime
        newest_src = 0.0
        for go_file in repo_root.rglob("*.go"):
            if go_file.name.endswith("_test.go"):
                continue
            m = go_file.stat().st_mtime
            if m > newest_src:
                newest_src = m
        if newest_src <= bin_mtime:
            return
        reason = "source newer than binary"
    else:
        reason = "binary missing"

    print(f"=== rebuilding {binary.name} ({reason}) ===", file=sys.stderr)
    sys.stderr.flush()
    binary.parent.mkdir(parents=True, exist_ok=True)
    result = subprocess.run(
        ["go", "build", "-o", str(binary), "./cmd/suneidotypes"],
        cwd=str(repo_root),
    )
    if result.returncode != 0:
        sys.exit(f"error: rebuild failed (go build exit {result.returncode})")


def run_binary(binary: str, request: dict, trace: bool = False) -> dict:
    """Pipe `request` to the binary as JSON and parse the JSON response.

    Captures stderr separately so the binary's own diagnostic stream
    (the colored [ERROR]/[WARNING] lines the plain CLI emits) does not
    contaminate the JSON payload on stdout.
    """
    request_json = json.dumps(request)
    print("=== request (stdin -> binary) ===")
    print(json.dumps(request, indent=6))
    env = os.environ.copy()
    if trace:
        env["GODEBUG"] = (env.get("GODEBUG", "") + ",inittrace=1").lstrip(",")
    proc = subprocess.run(
        [binary],
        input=request_json,
        text=True,
        capture_output=True,
        env=env,
    )
    if trace and proc.stderr:
        sys.stdout.flush()
        sys.stderr.write("=== trace (stderr) ===\n")
        sys.stderr.write(proc.stderr)
        if not proc.stderr.endswith("\n"):
            sys.stderr.write("\n")
        sys.stderr.flush()
    print("=== response (binary -> stdout) ===")
    try:
        print(json.dumps(json.loads(proc.stdout), indent=6))
    except json.JSONDecodeError:
        print(proc.stdout, end="" if proc.stdout.endswith("\n") else "\n")
    if proc.returncode != 0 and not proc.stdout.strip():
        msg = f"error: binary exited {proc.returncode}"
        if proc.stderr:
            msg += "\n" + proc.stderr.rstrip()
        sys.exit(msg)
    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        sys.exit(
            f"error: binary returned invalid JSON: {e}\n"
            f"--- stdout ---\n{proc.stdout}\n"
            f"--- stderr ---\n{proc.stderr}"
        )


def build_json_payload(
    path: str, class_name: str, method: str, response: dict
) -> dict:
    """Normalize the binary's response into a self-contained runner-level
    record. Always returns a dict with the same top-level keys regardless
    of method, so consumers don't have to branch on `method` to find the
    payload - the relevant slot under `result` is populated and the
    other is omitted.
    """
    results = response.get("result") or []
    first = results[0] if results else None

    result_block: dict[str, object] = {}
    if method == "TypeInfer":
        # first is {"methods": {m: {slot: type}}, "members": {name: type}}
        result_block["types"] = first if isinstance(first, dict) else {}
    elif method == "TypeAnnotate":
        result_block["annotated"] = first if isinstance(first, str) else ""

    diagnostics = response.get("diagnostics") or {"errors": [], "warnings": []}
    diagnostics.setdefault("errors", [])
    diagnostics.setdefault("warnings", [])

    return {
        "path": path,
        "class": class_name,
        "method": method,
        "ok": len(diagnostics["errors"]) == 0,
        "result": result_block,
        "diagnostics": diagnostics,
    }


def print_types(name: str, result: dict) -> None:
    """Render the TypeInfer payload: per-method slot types ($return, params,
    and locals, undifferentiated) plus the class's own data members."""
    methods = result.get("methods") or {}
    members = result.get("members") or {}

    print(f"=== types: {name} ===")
    if not methods and not members:
        print("  (nothing inferred)")
        return

    for method_name in sorted(methods):
        slots = methods[method_name]
        print(f"  {method_name}")
        if not slots:
            print("    (no inferred slots)")
            continue
        width = max(len(k) for k in slots)
        # $return first, then the remaining slots alphabetically
        for slot in sorted(slots, key=lambda k: (k != "$return", k)):
            print(f"    {slot.ljust(width)}  {slots[slot]}")

    if members:
        print("  members")
        width = max(len(k) for k in members)
        for member in sorted(members):
            print(f"    {member.ljust(width)}  {members[member]}")


def print_diagnostics(path: str, diagnostics: dict, use_color: bool) -> bool:
    """Render the diagnostics block. Returns True if any errors fired
    (so the caller can pick a non-zero exit code).
    """
    errors = diagnostics.get("errors", [])
    warnings = diagnostics.get("warnings", [])

    print()
    print("=== diagnostics ===")
    if not errors and not warnings:
        print(colored(ANSI_BOLD + ANSI_GREEN, "ok", use_color))
        return False

    # stderr is unbuffered while stdout is line-buffered, so without an
    # explicit flush the diagnostic lines below would race ahead of the
    # banner we just printed and land at the top of the terminal output.
    sys.stdout.flush()

    for d in errors:
        label = colored(ANSI_BOLD + ANSI_RED, "[ERROR]", use_color)
        print(
            f"{label} {path}:{d['line']}:{d['col']}: {d['method']}: {d['msg']}",
            file=sys.stderr,
        )
    for d in warnings:
        label = colored(ANSI_BOLD + ANSI_YELLOW, "[WARNING]", use_color)
        print(
            f"{label} {path}:{d['line']}:{d['col']}: {d['method']}: {d['msg']}",
            file=sys.stderr,
        )
    return bool(errors)


def main() -> int:
    parser = argparse.ArgumentParser(
        prog="runner.py",
        description="Run suneidotypes on a Suneido source file via the JSON protocol.",
    )
    parser.add_argument("path", help="Path to .ss / .suneido source file")
    parser.add_argument(
        "--annotate",
        action="store_true",
        help="Print the source spliced with inline type annotations",
    )
    parser.add_argument(
        "--name",
        help="Class name override (default: file stem)",
    )
    parser.add_argument(
        "--no-color",
        action="store_true",
        help="Disable ANSI color codes in diagnostics",
    )
    parser.add_argument(
        "--json",
        dest="emit_json",
        action="store_true",
        help="Emit a normalized JSON response on stdout and nothing else. "
             "Suitable for piping into other tools.",
    )
    parser.add_argument(
        "--trace",
        action="store_true",
        help="Run the binary with GODEBUG=inittrace=1 and forward its "
             "stderr so package init timings are visible.",
    )
    args = parser.parse_args()

    src_path = Path(args.path)
    if not src_path.is_file():
        sys.exit(f"error: not a file: {args.path}")
    src = src_path.read_text(encoding="utf-8")
    class_name = args.name or src_path.stem

    method = "TypeAnnotate" if args.annotate else "TypeInfer"
    request = {
        "method": method,
        "arguments": [{"name": class_name, "src": src}],
        "config": CONFIG,
    }

    response = run_binary(find_binary(), request, trace=args.trace)

    if args.emit_json:
        # Normalized runner-level shape. Differs from the binary's
        # raw response in three ways:
        #   1. Single-class results are unwrapped (no result[] array).
        #   2. The result is keyed by method type (`types` vs
        #      `annotated`) so consumers don't have to inspect `method`
        #      to know what shape they're getting.
        #   3. Added top-level path + ok so a downstream tool can
        #      treat the response as a self-contained record.
        payload = build_json_payload(args.path, class_name, method, response)
        json.dump(payload, sys.stdout, indent=2)
        sys.stdout.write("\n")
        return 0 if payload["ok"] else 1

    # NOTE: enabling colors when stdout is a pipe still works because
    # diagnostics go to stderr (which may be a tty even when stdout isn't).
    # The check covers the common case of `python runner.py ... > out.txt`.
    use_color = not args.no_color and sys.stderr.isatty()

    results = response.get("result", [])
    if method == "TypeInfer":
        if results:
            print_types(class_name, results[0])
    elif method == "TypeAnnotate":
        # results is a list aligned with the request's `arguments`. Today
        # runner.py only ever sends one class, but iterate so future
        # multi-class requests render in input order with a header per
        # class.
        for i, annotated in enumerate(results):
            if len(results) > 1:
                print(f"=== annotated source [{i}] ===")
            else:
                print("=== annotated source ===")
            print(annotated)

    had_errors = print_diagnostics(
        args.path, response.get("diagnostics", {}), use_color
    )
    return 1 if had_errors else 0


if __name__ == "__main__":
    sys.exit(main())
