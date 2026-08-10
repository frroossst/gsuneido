package annotations

import (
	"strings"

	"github.com/frroossst/SuneidoTypes/typeinfer/typealgebra"
)

type TypeSignature struct {
	Kind   string
	Prefix string
	Name   string
	Sig    string
}

type Set map[string][]Signature

type Param struct {
	Name       string
	Typ        typealgebra.DynType
	HasDefault bool
	Inferred   bool   // Typ came from RequirementPass usage inference, not an annotation
	Why        string // Inferred only: root callee whose contract the demand chains back to
}

type Signature struct {
	Receiver typealgebra.DynType // nil = unspecified
	Params   []Param
	AtParam  bool
	Returns  typealgebra.DynType
}

var methodPrefixReceiver = map[string]typealgebra.DynType{
	"ob":     typealgebra.TObject,
	"string": typealgebra.TString,
	"num":    typealgebra.TNumber,
	"int":    typealgebra.TNumber, // SuInt methods are Number methods
	"date":   typealgebra.TDate,
	"class":  typealgebra.TClass,
	"func":   typealgebra.TFunction,
	"seq":    typealgebra.TSequence, // TSequence <: TObject: seq methods add to the inherited Object set
}

var staticPrefixClass = map[string]string{
	"dateStatic": "Date",
	"db":         "Database",
	"ftsearch":   "Ftsearch",
	"lruStatic":  "LruCache",
	"opgp":       "OpenPGP",
	"pe":         "PdfEncrypt",
	"rnd":        "Random",
	"sqs":        "Query",
	"suneido":    "Suneido",
	"thread":     "Thread",
	"zlib":       "Zlib",
}

// gsuneido emits db-record methods under this prefix.
const recordPrefix = "record"

// precedence, first registration wins: imported interpreter sigs, then overlay gap-fill, then record methods (which must never shadow Object); second return lists every dropped input with its reason
func Load(imported []TypeSignature) (Set, []string) {
	l := &loader{set: Set{}, seen: map[sigKey]bool{}}
	l.loadImported(imported)
	l.applyOverlay()
	l.loadRecordMethodsAsObject(imported)
	return l.set, l.skipped
}

type loader struct {
	set     Set
	seen    map[sigKey]bool
	skipped []string
}

type sigKey struct {
	name string
	recv string
}

func (l *loader) register(name string, s Signature) {
	k := sigKey{name: name, recv: receiverKey(s.Receiver)}
	if l.seen[k] {
		return
	}
	l.seen[k] = true
	l.set[name] = append(l.set[name], s)
}

func (l *loader) loadImported(sigs []TypeSignature) {
	for _, ts := range sigs {
		switch ts.Kind {
		case "free":
			if s, ok := l.parseImportedSig(ts, nil); ok {
				l.register(ts.Name, s)
			}
		case "method":
			l.loadImportedMethod(ts)
		case "static":
			l.loadImportedStatic(ts)
		}
	}
}

func (l *loader) loadImportedMethod(ts TypeSignature) {
	if ts.Prefix == recordPrefix {
		return
	}
	recv, ok := methodPrefixReceiver[ts.Prefix]
	if !ok {
		l.skipped = append(l.skipped, "method "+ts.Prefix+"."+ts.Name)
		return
	}
	if s, ok := l.parseImportedSig(ts, recv); ok {
		l.register(ts.Name, s)
	}
}

func (l *loader) loadImportedStatic(ts TypeSignature) {
	class, ok := staticPrefixClass[ts.Prefix]
	if !ok {
		l.skipped = append(l.skipped, "static "+ts.Prefix+"."+ts.Name)
		return
	}
	if s, ok := l.parseImportedSig(ts, nil); ok {
		l.register(class+"."+ts.Name, s)
	}
}

func (l *loader) loadRecordMethodsAsObject(sigs []TypeSignature) {
	for _, ts := range sigs {
		if ts.Kind != "method" || ts.Prefix != recordPrefix {
			continue
		}
		if s, ok := l.parseImportedSig(ts, typealgebra.TObject); ok {
			l.register(ts.Name, s)
		}
	}
}

func receiverKey(r typealgebra.DynType) string {
	if r == nil {
		return ""
	}
	return r.String()
}

func (l *loader) parseImportedSig(ts TypeSignature, recv typealgebra.DynType) (Signature, bool) {
	s, err := parseSig(sanitizeSig(ts.Sig))
	if err != nil {
		l.skipped = append(l.skipped,
			ts.Kind+" "+ts.Prefix+"."+ts.Name+": "+err.Error())
		return Signature{}, false
	}
	s.Receiver = recv
	if s.Returns == nil {
		// nil is not a valid type, so turn to TUnknown
		s.Returns = typealgebra.TUnknown
	}
	return s, true
}

func sanitizeSig(sig string) string {
	return strings.ReplaceAll(sig, "nil", "'nil'")
}
