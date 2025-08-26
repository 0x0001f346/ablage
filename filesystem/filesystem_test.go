package filesystem

import (
	"math"
	"testing"
)

func Test_sanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "1",
			input: "test.png",
			want:  "test.png",
		},
		{
			name:  "2",
			input: "/tmp/test.png",
			want:  "test.png",
		},
		{
			name:  "3",
			input: "../../etc/passwd",
			want:  "passwd",
		},
		{
			name:  "4",
			input: "",
			want:  "upload.bin",
		},
		{
			name:  "5",
			input: "übergrößé.png",
			want:  "uebergroess.png",
		},
		{
			name:  "6",
			input: "my cool file!!.txt",
			want:  "my_cool_file.txt",
		},
		{
			name:  "7",
			input: "so  many   spaces.txt",
			want:  "so_many_spaces.txt",
		},
		{
			name:  "8",
			input: "/tmp/abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz.txt",
			want:  "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeFilename(tt.input); got != tt.want {
				t.Errorf("\nsanitizeFilename()\nname: %v\nwant: %v\ngot:  %v", tt.name, tt.want, got)
			}
		})
	}
}

func Test_getHumanReadableSize(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{
			name:  "1",
			input: 7,
			want:  "7 Bytes",
		},
		{
			name:  "2",
			input: 7 * int64(math.Pow(10, 3)),
			want:  "6.8 KB",
		},
		{
			name:  "3",
			input: 7 * int64(math.Pow(10, 6)),
			want:  "6.7 MB",
		},
		{
			name:  "4",
			input: 7 * int64(math.Pow(10, 9)),
			want:  "6.5 GB",
		},
		{
			name:  "5",
			input: 7 * int64(math.Pow(10, 12)),
			want:  "6.4 TB",
		},
		{
			name:  "6",
			input: 7 * int64(math.Pow(10, 15)),
			want:  "6.2 PB",
		},
		{
			name:  "7",
			input: 7 * int64(math.Pow(10, 18)),
			want:  "6.1 EB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHumanReadableSize(tt.input); got != tt.want {
				t.Errorf("\ngetHumanReadableSize()\nname: %v\nwant: %v\ngot:  %v", tt.name, tt.want, got)
			}
		})
	}
}
