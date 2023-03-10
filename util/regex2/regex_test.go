// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package regex2

import (
	"fmt"
	"testing"

	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/ptest"
)

func TestPlay(t *testing.T) {
	s := "abc"
	pat := Compile(`ab$`)
	fmt.Println(pat)
	var cap Captures
	fmt.Println(">>>", pat.Match(s, &cap), cap[0])
	cap.Print(s)
}

func (c *Captures) Print(s string) {
	// fmt.Println(cap)
	for i := 0; i < 20; i += 2 {
		if i == 0 || c[i] > 0 || c[i+1] > 0 {
			fmt.Printf("%d: %q\n", i/2, s[c[i]:c[i+1]])
		}
	}
}

// func TestGoRegexp(t *testing.T) {
// 	pat := regexp.MustCompile(`(?m)^[^x].*$`)
// 	cap := pat.FindStringSubmatch("xyz\r\n\r\nxyz")
// 	fmt.Printf("%#v\n", cap)
// }

func TestCapture(t *testing.T) {
	test := func(str, pat string, expected ...string) {
		t.Helper()
		var cap Captures
		Compile(pat).Match(str, &cap)
		// cap.Print(str)
		for i, e := range expected {
			assert.T(t).This(str[cap[i+2]:cap[i+3+1]]).Is(e)
		}
	}
	test("abcd", "(.+)(.+)", "abc", "d")
	test("abcd", "(.+?)(.+)", "a", "bcd")
	test("abcd", "(.*)(.*)", "abcd", "")
	test("abcd", "(.*?)(.*)", "", "abcd")
}

func TestMatch(t *testing.T) {
	match := func(str string, pat string, expected bool) {
		t.Helper()
		// fmt.Printf("%q =~ %q -> %v\n", str, pat, expected)
		assert.T(t).This(Compile(pat).Match(str, nil)).Is(expected)
	}
	matchAll := func(str string, pat string, expected bool) {
		t.Helper()
		pat = `^` + pat + `$`
		match(str, pat, expected)
	}
	// literal
	matchAll("a", "a", true)
	match("a", "a", true)
	match("abc", "b", true)
	match("abc", `^b`, false)
	match("abc", `^ab`, true)
	match("abc", `bc$`, true)
	match("abc", `b$`, false)
	match("abc", `^abc$`, true)
	match("abc", `^b$`, false)
	matchAll("a", "", false)
	matchAll("", "a", false)
	matchAll("a", "b", false)

	// one pass
	matchAll("a", ".", true)
	matchAll("", ".", false)
	matchAll("abc", ".bc", true)
	matchAll("abc", ".bx", false)

	// full
	matchAll("a", "a|b", true)
	matchAll("b", "a|b", true)
	matchAll("", "a|b", false)
	matchAll("c", "a|b", false)
	matchAll("", "a?", true)
	matchAll("a", "a?", true)
}

func TestCompile(t *testing.T) {
	test := func(rx, expected string) {
		t.Helper()
		assert.T(t).This(Compile(rx).String()).Like(expected)
	}
	test("xyz",
		`0: LiteralSubstr "xyz"`)
	test(`^xyz`,
		`0: LiteralPrefix "xyz"`)
	test(`xyz$`,
		`0: LiteralSuffix "xyz"`)
	test(`^xyz$`,
		`0: LiteralEqual "xyz"`)
	test(`.`,
		`0: AnyNotNL
		1: DoneSave1`)
	test(`^.`,
		`0: OnePass
		1: StrStart
		2: AnyNotNL
		3: DoneSave1`)
	test("a|b",
		`0: SplitFirst 8
		3: Char a
		5: Jump 10
		8: Char b
		10: DoneSave1`)
	test("a|b|c",
		`0: SplitFirst 8
		3: Char a
		5: Jump 18
		8: SplitFirst 16
		11: Char b
		13: Jump 18
		16: Char c
		18: DoneSave1`)
	test("ab?c",
		`0: Prefix "a"
		3: Char a
		5: SplitLast 10
		8: Char b
		10: Char c
		12: DoneSave1`)
	test("ab+c",
		`0: Prefix "ab"
		4: Char a
		6: Char b
		8: SplitFirst 6
		11: Char c
		13: DoneSave1`)
	test("ab*c",
		`0: Prefix "a"
		3: Char a
		5: SplitLast 13
		8: Char b
		10: Jump 5
		13: Char c
		15: DoneSave1`)
}

func BenchmarkOnePass(b *testing.B) {
	pat := Compile("abc")
	for i := 0; i < b.N; i++ {
		pat.Match("abc", nil)
	}
}

// ptest support ---------------------------------------------------------------

func TestPtest(t *testing.T) {
	if !ptest.RunFile("regex.test") {
		t.Fail()
	}
}

func TestPtest2(t *testing.T) {
	result := ptMatch([]string{"a", "a?", "a"}, nil)
	fmt.Println(result)
}

// pt_match is a ptest for matching
// simple usage is two arguments, string and pattern
// an optional third argument can be "false" for matches that should fail
// or additional arguments can specify expected \0, \1, ...
func ptMatch(args []string, _ []bool) bool {
	// fmt.Println(args)
	s := args[0]
	pat := Compile("(?m)" + args[1])
	var cap Captures
	result := pat.Match(s, &cap)
	if len(args) > 2 {
		if args[2] == "false" {
			result = !result
		} else {
			for i, e := range args[2:] {
				p := s[cap[i*2]:cap[i*2+1]]
				result = result && (e == p)
			}
		}
	}
	return result
}

var _ = ptest.Add("regex_match", ptMatch)

/*
// pt_replace is a ptest for regex replace
func ptReplace(args []string, _ []bool) bool {
	s := args[0]
	pat := Compile(args[1])
	rep := args[2]
	expected := args[3]
	var cap Captures
	result := pat.FirstMatch(s, &cap)
	if !result {
		return false
	}
	r := Replacement(s, rep, &cap)
	pos, end := cap[0].Range()
	t := s[:pos] + r + s[end:]
	if t != expected {
		fmt.Println("\t     got:", t, "\n\texpected:", expected)
		return false
	}
	return true
}

var _ = ptest.Add("regex_replace", ptReplace)
*/
