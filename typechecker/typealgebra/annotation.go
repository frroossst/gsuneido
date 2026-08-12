package typealgebra

import (
	"fmt"
	"sort"
	"strings"
)

// This file is the annotation language: how a type is written, read back, and
// found in source. Everything about the surface syntax lives here, so changing
// how annotations look is a change to this file alone. The lattice itself is in
// typealgebra.go, and where an annotation gets spliced into a file is debug's.

// armSep separates the alternatives of a union, with no spaces around it.
const armSep = "|"

// dirtyArm marks a union with an unknown alternative. It is printable but not
// writable: no annotation can declare it.
const dirtyArm = "?"

// primitiveNames are the spellings of the primitives, and the one way a type is
// written anywhere: annotations, inferred-type output, and diagnostics. They
// are lowercase because capitalization means a class name. Note these are not
// Suneido's own `Type(x)` strings, which stay capitalized.
var primitiveNames = [...]string{
	"unknown", "void", "boolean", "false", "true", "number", "string", "date",
	"function", "block", "class", "object", "sequence",
}

// primitiveByName inverts primitiveNames, so writing and reading a type name
// cannot drift apart.
var primitiveByName = func() map[string]Primitive {
	m := make(map[string]Primitive, len(primitiveNames))
	for i, n := range primitiveNames {
		m[n] = Primitive(i)
	}
	return m
}()

// typeAliases are accepted spellings with no rendering of their own. Suneido's
// Record is an Object as far as the type lattice is concerned, but a source
// that says `record` keeps saying it.
var typeAliases = map[string]Primitive{
	"record": TObject,
}

// ParseName resolves a written builtin type name, matched case-insensitively.
// It is the inverse of Primitive.String, plus the aliases.
func ParseName(name string) (Primitive, bool) {
	lower := strings.ToLower(name)
	if p, ok := primitiveByName[lower]; ok {
		return p, true
	}
	p, ok := typeAliases[lower]
	return p, ok
}

// CanonicalName returns the spelling of a builtin type name, matched
// case-insensitively, and reports whether name is a builtin at all. Aliases are
// not canonicalized: `record` is left as written rather than turned into
// `object`, since only the author knows which they meant.
func CanonicalName(name string) (string, bool) {
	if p, ok := primitiveByName[strings.ToLower(name)]; ok {
		return p.String(), true
	}
	return name, false
}

// ParseAnnotation reads an annotation into a type. A name is nominal because it
// is not a builtin, not because it is capitalized: `String` and `string` are
// one type, while `Foo` is a class. On error it still returns the best type it
// could make, with the unreadable alternatives left unknown.
func ParseAnnotation(s string) (DynType, error) {
	if s == "" {
		return TUnknown, nil
	}
	var result DynType
	var errs []string
	for _, raw := range strings.Split(s, armSep) {
		name := strings.TrimSpace(raw)
		if name == "" {
			errs = append(errs, "empty type alternative")
			continue
		}
		t, err := parseArm(name)
		if err != nil {
			errs = append(errs, err.Error())
			t = TUnknown
		}
		if result == nil {
			result = t
		} else {
			result = U(result, t)
		}
	}
	if result == nil {
		result = TUnknown
	}
	if len(errs) > 0 {
		return result, fmt.Errorf("annotation %q: %s", s, strings.Join(errs, "; "))
	}
	return result, nil
}

func parseArm(name string) (DynType, error) {
	if t, ok := ParseName(name); ok {
		return t, nil
	}
	if name[0] >= 'A' && name[0] <= 'Z' {
		return Instance{Class: name}, nil
	}
	return TUnknown, fmt.Errorf("unknown type name %q", name)
}

// NativeAnnotation is String restricted to what native `:type` syntax can
// express, reporting false for a type that has to be written as a comment.
func NativeAnnotation(ty DynType) (string, bool) {
	switch t := ty.(type) {
	case Primitive:
		if t == TUnknown {
			return "", false
		}
		return t.String(), true
	case Instance:
		if t.Class == "" {
			return "", false
		}
		name, _ := CanonicalName(t.Class)
		return name, true
	case Union:
		if t.IsDirty {
			return "", false // the ? alternative is not expressible
		}
		parts := make([]string, 0, len(t.Types))
		for _, m := range t.Types {
			s, ok := NativeAnnotation(m)
			if !ok {
				return "", false
			}
			parts = append(parts, s)
		}
		if len(parts) == 0 {
			return "", false
		}
		return joinArms(parts), true
	default:
		return "", false
	}
}

// CanonicalAnnotation respells a declared annotation without changing what it
// means, so a source that says `String | False` can be rewritten to the one
// spelling everything else uses.
func CanonicalAnnotation(decl string) string {
	parts := strings.Split(decl, armSep)
	for i, p := range parts {
		name, _ := CanonicalName(strings.TrimSpace(p))
		parts[i] = name
	}
	return strings.Join(parts, armSep)
}

// joinArms sorts so that semantically equal unions print identically whatever
// order they were built in. "?" sorts last on its own, after every name.
func joinArms(parts []string) string {
	sorted := append([]string(nil), parts...)
	sort.Strings(sorted)
	return strings.Join(sorted, armSep)
}

// AnnotationSpan returns the end of the `: type | type` annotation written at
// from, so it can be replaced with its canonical spelling. It reports false
// when the annotation is not laid out on one line, leaving that source alone.
func AnnotationSpan(src string, from int) (int, bool) {
	i := skipBlank(src, from)
	if i >= len(src) || src[i] != ':' {
		return 0, false
	}
	i++
	for {
		i = skipBlank(src, i)
		j := i
		for j < len(src) && isIdentChar(src[j]) {
			j++
		}
		if j == i {
			return 0, false // a colon with no type name after it
		}
		i = j
		if k := skipBlank(src, i); k < len(src) && src[k] == armSep[0] {
			i = k + 1
			continue
		}
		return i, true
	}
}

func skipBlank(src string, i int) int {
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	return i
}

func isIdentChar(c byte) bool {
	return c == '_' || c == '?' || c == '!' ||
		('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}
