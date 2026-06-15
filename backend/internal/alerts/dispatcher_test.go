package alerts

import (
	"sort"
	"testing"
)

func toSet(ids []string) map[string]struct{} {
	s := map[string]struct{}{}
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

func TestDiff(t *testing.T) {
	cases := []struct {
		name        string
		prev, curr  []string
		wantEntered []string
		wantExited  []string
	}{
		{"no change", []string{"a", "b"}, []string{"a", "b"}, nil, nil},
		{"pure entry", []string{}, []string{"a"}, []string{"a"}, nil},
		{"pure exit", []string{"a"}, []string{}, nil, []string{"a"}},
		{"swap", []string{"a"}, []string{"b"}, []string{"b"}, []string{"a"}},
		{"mixed", []string{"a", "b", "c"}, []string{"b", "d"}, []string{"d"}, []string{"a", "c"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entered, exited := Diff(toSet(tc.prev), toSet(tc.curr))
			sort.Strings(entered)
			sort.Strings(exited)
			if !equal(entered, tc.wantEntered) {
				t.Fatalf("entered got %v want %v", entered, tc.wantEntered)
			}
			if !equal(exited, tc.wantExited) {
				t.Fatalf("exited got %v want %v", exited, tc.wantExited)
			}
		})
	}
}

func equal(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
