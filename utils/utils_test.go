package utils

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Les misérables", "Les miserables"},
		{" Kagerô-za", "Kagero-za"},
		{" Méliès: Fairy Tales in Color", "Melies Fairy Tales in Color"},
		{" Love (2D  +    3D) ", "Love (2D + 3D)"},
		{" F/X: The Usual      Suspects  ", "FX The Usual Suspects"},
		{"Another / Movie /      Test", "Another Movie Test"},
		{"", ""},
	}

	for _, c := range cases {
		got, _ := Normalize(c.in)
		if got != c.want {
			t.Errorf("Normalize(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRemoveURLParams(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://defret.in/log/dark-theme-on-ruby-on-rails", "https://defret.in/log/dark-theme-on-ruby-on-rails"},
		{"https://www.youtube.com/watch?v=LvgVSSpwND8", "https://www.youtube.com/watch"},
		{"https://i1.wp.com/caps.pictures/198/0-airplane/full/airplane-movie-screencaps.com-1.jpg?strip=all",
			"https://i1.wp.com/caps.pictures/198/0-airplane/full/airplane-movie-screencaps.com-1.jpg"},
		{"https://film-grab.com/wp-content/uploads/photo-gallery/imported_from_media_libray/10cloverfield001.jpg?bwg=1546957525",
			"https://film-grab.com/wp-content/uploads/photo-gallery/imported_from_media_libray/10cloverfield001.jpg"},
		{"https://stillsfrmfilms.files.wordpress.com/2014/02/012.jpg?w=150&amp;h=64",
			"https://stillsfrmfilms.files.wordpress.com/2014/02/012.jpg"},
		{"", ""},
	}

	for _, c := range cases {
		got := RemoveURLParams(c.in)
		if got != c.want {
			t.Errorf("RemoveURLParams(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
