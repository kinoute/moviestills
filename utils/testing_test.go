package utils

import (
	"testing"
)

func TestGetHTMLCode(t *testing.T) {
	cases := []struct {
		in string
	}{
		{"https://defret.in"},
		{"https://github.com"},
		{"https://whattheshot.com"},
		{"https://whatthemovie.com"},
	}

	for _, c := range cases {
		got := GetHTMLCode(c.in)
		title := got.Find("title").Text()
		if title == "" {
			t.Errorf("getHTMLCode(%q) == null, expected not empty", c.in)
		}
	}
}
