// Copyright (C) 2011 Florian Weimer <fw@deneb.enyo.de>

package pcre

import (
	"testing"
)

func TestCompile(t *testing.T) {
	var check = func (p string, groups int) {
		re, err := Compile(p, 0)
		if err != nil {
			t.Error(p, err)
		}
		if g := re.Groups(); g != groups {
			t.Error(p, g)
		}
	}
	check("",0 )
	check("^", 0)
	check("^$", 0)
	check("()", 1)
	check("(())", 2)
	check("((?:))", 1)
}

func TestCompileFail(t *testing.T) {
	var check = func (p, msg string, off int) {
		_, err := Compile(p, 0)
		switch {
		case err == nil:
			t.Error(p)
		case err.Message != msg:
			t.Error(p, "Message", err.Message)
		case err.Offset != off:
			t.Error(p, "Offset", err.Offset)
		}
	}
	check("(", "missing )", 1)
	check("\\", "\\ at end of pattern", 1)
	check("abc\\", "\\ at end of pattern", 4)
	check("abc\000", "NUL byte in pattern", 3)
	check("a\000bc", "NUL byte in pattern", 1)
}

func strings(b [][]byte) (r []string) {
	r = make([]string, len(b))
	for i, v := range b {
		r[i] = string(v)
	} 
	return
}

func equal(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}
	for i, lv := range l {
		if lv != r[i] {
			return false
		}
	}
	return true
}

func checkmatch1(t *testing.T, dostring bool, m *Matcher,
	pattern, subject string, args ...interface{}) {
	re := MustCompile(pattern, 0)
	var prefix string
	if dostring {
		if m == nil {
			m = re.MatcherString(subject, 0)
		} else {
			m.ResetString(re, subject, 0)
		}
		prefix = "string"
	} else {
		if m == nil {
			m = re.Matcher([]byte(subject), 0)
		} else {
			m.Reset(re, []byte(subject), 0)
		}
		prefix = "[]byte"
	}
	if len(args) == 0 {
		if m.Matches() {
			t.Error(prefix, pattern, subject, "!Matches")
		}
	} else {
		if !m.Matches() {
			t.Error(prefix, pattern, subject, "Matches")
			return
		}
		if m.Groups() != len(args) - 1 {
			t.Error(prefix, pattern, subject, "Groups", m.Groups())
			return
		}
		for i, arg := range args {
			if s, ok := arg.(string); ok {
				if !m.Present(i) {
					t.Error(prefix, pattern, subject,
						"Present", i)

				}
				if g := string(m.Group(i)); g != s {
					t.Error(prefix, pattern, subject,
						"Group", i, g, "!=", s)
				}
				if g := m.GroupString(i); g != s {
					t.Error(prefix, pattern, subject,
						"GroupString", i, g, "!=", s)
				}
			} else {
				if m.Present(i) {
					t.Error(prefix, pattern, subject,
						"!Present", i)
				}
			}
		}
	}
}

func TestMatcher(t *testing.T) {
	var m Matcher
	check := func(pattern, subject string, args ...interface{}) {
		checkmatch1(t, false, nil, pattern, subject, args...)
		checkmatch1(t, true, nil, pattern, subject, args...)
		checkmatch1(t, false, &m, pattern, subject, args...)
		checkmatch1(t, true, &m, pattern, subject, args...)
	}

	check(`^$`, "", "")
	check(`^abc$`, "abc", "abc")
	check(`^(X)*ab(c)$`, "abc", "abc", nil, "c")
	check(`^(X)*ab()c$`, "abc", "abc", nil, "")
	check(`^.*$`, "abc", "abc")
	check(`^.*$`, "a\000c", "a\000c")
	check(`^(.*)$`, "a\000c", "a\000c", "a\000c")
}

func TestCaseless(t *testing.T) {
	m := MustCompile("abc", CASELESS).MatcherString("Abc", 0)
	if !m.Matches() {
		t.Error("CASELESS")
	}
	m = MustCompile("abc", 0).MatcherString("Abc", 0)
	if m.Matches() {
		t.Error("!CASELESS")
	}
}

func TestNamed(t *testing.T) {
	m := MustCompile("(?<L>a)(?<M>X)*bc(?<DIGITS>\\d*)", 0).
		MatcherString("abc12", 0)
	if !m.Matches() {
		t.Error("Matches")
	}
	if !m.NamedPresent("L") {
		t.Error("NamedPresent(\"L\")")
	}
	if m.NamedPresent("M") {
		t.Error("NamedPresent(\"M\")")
	}
	if !m.NamedPresent("DIGITS") {
		t.Error("NamedPresent(\"DIGITS\")")
	}
	if "12" != m.NamedString("DIGITS") {
		t.Error("NamedString(\"DIGITS\")")
	}
}

func TestFindIndex(t *testing.T) {
	re := MustCompile("bcd", 0)
	i := re.FindIndex([]byte("abcdef"), 0)
	if i[0] != 1 {
		t.Error("FindIndex start", i[0])
	}
	if i[1] != 4 {
		t.Error("FindIndex end", i[1])
	}
}

func TestReplaceAll(t *testing.T) {
	re := MustCompile("foo", 0)
	// Don't change at ends.
	result := re.ReplaceAll([]byte("I like foods."), []byte("car"), 0)
	if string(result) != "I like cards." {
		t.Error ("ReplaceAll", result)
	}
	// Change at ends.
	result = re.ReplaceAll([]byte("food fight fools foo"), []byte("car"), 0)
	if string(result) != "card fight carls car" {
		t.Error("ReplaceAll2", result)
	}
}
