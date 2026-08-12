package annotations_test

import (
	"strings"
	"testing"

	"github.com/apmckinlay/gsuneido/typechecker/annotations"
	"github.com/apmckinlay/gsuneido/typechecker/typealgebra"
)

// sigFor returns the entry under name with an Object receiver, or nil.
func sigFor(set annotations.Set, name string) *annotations.Signature {
	for i, s := range set[name] {
		if s.Receiver == typealgebra.TObject {
			return &set[name][i]
		}
	}
	return nil
}

func TestLoadImportedBeatsOverlay(t *testing.T) {
	set, _ := annotations.Load([]annotations.TypeSignature{
		{Kind: "method", Prefix: "ob", Name: "Map", Sig: "(f) :number"},
	})
	s := sigFor(set, "Map")
	if s == nil {
		t.Fatal("Map on Object missing")
	}
	if s.Returns != typealgebra.TNumber {
		t.Fatalf("imported signature lost to overlay: returns %v, want Number", s.Returns)
	}
	if len(set["Map"]) != 2 {
		// the overlay's String.Map must still load - different receiver
		t.Fatalf("want 2 Map entries (Object imported + String overlay), got %d", len(set["Map"]))
	}
}

func TestLoadOverlayFillsGaps(t *testing.T) {
	set, _ := annotations.Load(nil)
	s := sigFor(set, "DeleteIf")
	if s == nil {
		t.Fatal("overlay-only method DeleteIf missing with empty import")
	}
	if s.Returns != typealgebra.TObject {
		t.Fatalf("DeleteIf returns %v, want Object", s.Returns)
	}
}

func TestLoadRecordMethodsFillGapsOnly(t *testing.T) {
	set, _ := annotations.Load([]annotations.TypeSignature{
		// collides with the overlay's Object Map - must be dropped
		{Kind: "method", Prefix: "record", Name: "Map", Sig: "(x) :string"},
		// record-only - must be added on an Object receiver
		{Kind: "method", Prefix: "record", Name: "AttachRule", Sig: "(rule) :object"},
	})
	if s := sigFor(set, "Map"); s == nil || s.Returns != typealgebra.TObject {
		t.Fatalf("record Map shadowed the Object Map: %+v", s)
	}
	if s := sigFor(set, "AttachRule"); s == nil {
		t.Fatal("record-only AttachRule not registered on Object")
	}
}

func TestLoadStaticAndFreeAndSkipped(t *testing.T) {
	set, skipped := annotations.Load([]annotations.TypeSignature{
		{Kind: "static", Prefix: "dateStatic", Name: "Now", Sig: "() :date"},
		{Kind: "free", Prefix: "", Name: "Print", Sig: "(@args)"},
		{Kind: "method", Prefix: "mystery", Name: "Zap", Sig: "()"},
		{Kind: "method", Prefix: "ob", Name: "Broken", Sig: "(((("},
	})
	if len(set["Date.Now"]) != 1 {
		t.Fatal("static dateStatic.Now not registered as Date.Now")
	}
	if len(set["Print"]) != 1 || !set["Print"][0].AtParam {
		t.Fatalf("free Print not registered as @args variadic: %+v", set["Print"])
	}
	var unmapped, unparsable bool
	for _, s := range skipped {
		if s == "method mystery.Zap" {
			unmapped = true
		}
		if strings.HasPrefix(s, "method ob.Broken:") {
			unparsable = true
		}
	}
	if !unmapped || !unparsable {
		t.Fatalf("skip reporting incomplete: %v", skipped)
	}
}

func TestLoadDeterministic(t *testing.T) {
	in := []annotations.TypeSignature{
		{Kind: "method", Prefix: "ob", Name: "Map", Sig: "(f) :number"},
		{Kind: "method", Prefix: "string", Name: "Zed", Sig: "(x) :string"},
	}
	a, _ := annotations.Load(in)
	b, _ := annotations.Load(in)
	if len(a) != len(b) {
		t.Fatalf("nondeterministic load: %d vs %d names", len(a), len(b))
	}
	for name, sa := range a {
		sb := b[name]
		if len(sa) != len(sb) {
			t.Fatalf("nondeterministic entry count for %s", name)
		}
		for i := range sa {
			if sa[i].AtParam != sb[i].AtParam || len(sa[i].Params) != len(sb[i].Params) {
				t.Fatalf("nondeterministic signature for %s[%d]", name, i)
			}
		}
	}
}
