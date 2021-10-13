package utils

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"Les misérables", "Les miserables"},
		{" Kagerô-za", "Kagero-za"},
		{" Méliès: Fairy Tales in Color", "Melies：Fairy Tales in Color"},
		{" Love (2D  +    3D) ", "Love (2D + 3D)"},
		{" F/X: The Usual      Suspects  ", "F X：The Usual Suspects"},
		{"Another / Movie /      Test", "Another Movie Test"},
		{"", ""},
	}

	for _, c := range cases {
		got, _ := Normalize(c.in)
		if got != c.expected {
			t.Errorf("Normalize(%q) == %q, expected %q", c.in, got, c.expected)
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
	type args struct {
		s      string
		length int
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "normal test",
			args: args{
				s:      "你好 hello",
				length: 8,
			},
			expected: "你好 hello",
		},
		{
			name: "truncated test",
			args: args{
				s:      "你好 hello",
				length: 6,
			},
			expected: "你好 ...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LimitLength(tt.args.s, tt.args.length); got != tt.expected {
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
	type args struct {
		moviePath string
	}
	cases := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "wrong path to create folder",
			args: args{
				moviePath: "/....$x/dev/null/root/8 Mile$$$\\..111@:£*\n",
			},
			expected: "",
		},
	}

	for _, c := range cases {
		got, _ := CreateFolder(c.args.moviePath)
		// if got is nil, there was an error during creation of the folder
		if got != c.expected {
			t.Errorf("CreateFolder(%q) == %q, expected %q", c.args.moviePath, got, c.expected)
		}
	}
}

func TestMD5(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "normal test",
			args: args{
				text: "123456",
			},
			expected: "e10adc3949ba59abbe56e057f20f883e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.args.text); got != tt.expected {
				t.Errorf("MD5() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestSaveImage(t *testing.T) {
	type args struct {
		moviePath   string
		movieName   string
		rawFileName string
		body        []byte
		toHash      bool
	}
	cases := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "can't save image in protected folder",
			args: args{
				moviePath:   "/....$x/dev/null/root/8 Mile$$$\\..111@:£*\n",
				movieName:   "8 Mile",
				rawFileName: "screen1.jpg",
				body:        []byte{96, 23},
				toHash:      true,
			},
			expected: "open /....$x/dev/null/root/8 Mile$$$\\..111@:£*\n/ba2c2d7263eff7f1c6cec59a018d27cc.jpg: no such file or directory",
		},
	}

	for _, c := range cases {
		got := SaveImage(c.args.moviePath, c.args.movieName, c.args.rawFileName, c.args.body, c.args.toHash)
		if got.Error() != c.expected {
			t.Errorf("SaveImage(%v) == %q, expected %q", c.args, got, c.expected)
		}
	}
}
