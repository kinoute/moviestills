package utils

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		in       string
		expected string
		wantErr  bool
	}{
		{"Les misérables", "Les miserables", false},
		{" Kagerô-za", "Kagero-za", false},
		{" Méliès: Fairy Tales in Color", "Melies：Fairy Tales in Color", false},
		{" Love (2D  +    3D) ", "Love (2D + 3D)", false},
		{" F/X: The Usual      Suspects  ", "F X：The Usual Suspects", false},
		{"Another / Movie /      Test", "Another Movie Test", false},
		{"", "", true},                    // Empty result error
		{"../../../etc/passwd", "", true}, // Path traversal
		{"movie/../secret", "", true},     // Path traversal
		{"Normal..Movie", "", true},       // Contains ..
	}

	for _, c := range cases {
		got, err := Normalize(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("Normalize(%q) expected error, got nil", c.in)
			}
		} else {
			if err != nil {
				t.Errorf("Normalize(%q) unexpected error: %v", c.in, err)
			}
			if got != c.expected {
				t.Errorf("Normalize(%q) == %q, expected %q", c.in, got, c.expected)
			}
		}
	}
}

func TestRemoveURLParams(t *testing.T) {
	cases := []struct {
		in       string
		expected string
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
		if got != c.expected {
			t.Errorf("RemoveURLParams(%q) == %q, expected %q", c.in, got, c.expected)
		}
	}
}

func TestLimitLength(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		length   int
		expected string
	}{
		{
			name:     "normal test",
			s:        "你好 hello",
			length:   8,
			expected: "你好 hello",
		},
		{
			name:     "truncated test",
			s:        "你好 hello",
			length:   6,
			expected: "你好 ...",
		},
		{
			name:     "unlimited",
			s:        "any string",
			length:   0,
			expected: "any string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LimitLength(tt.s, tt.length); got != tt.expected {
				t.Errorf("LimitLength() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestRemoveDisallowedChars(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"hello/world", "hello world"},
		{"hello:world", "hello：world"},
		{"2001|2020", "2001-2020"},
		{"", ""},
	}

	for _, c := range cases {
		got := RemoveDisallowedChars(c.in)
		if got != c.expected {
			t.Errorf("RemoveDisallowedChars(%q) == %q, expected %q", c.in, got, c.expected)
		}
	}
}

func TestCreateFolder(t *testing.T) {
	cases := []struct {
		name      string
		moviePath string
		wantErr   bool
	}{
		{
			name:      "invalid path",
			moviePath: "/....$x/dev/null/root/8 Mile$$$\\..111@:£*\n",
			wantErr:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := CreateFolder(c.moviePath)
			if c.wantErr {
				if err == nil && got != "" {
					t.Errorf("CreateFolder(%q) expected error or empty result, got %q", c.moviePath, got)
				}
			}
		})
	}
}
