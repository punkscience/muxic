package dedup_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"muxic/pkg/dedup"
)

func write(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func pattern(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill + byte(i%7)
	}
	return b
}

func find(t *testing.T, dir string, exts ...string) dedup.Result {
	t.Helper()
	r, err := dedup.Find(dir, dedup.Options{Extensions: exts, Workers: 4, Warn: func(err error) { t.Error(err) }})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func names(r dedup.Result) [][]string {
	var out [][]string
	for _, g := range r.Groups {
		var n []string
		for _, f := range g {
			n = append(n, filepath.Base(f.Path))
		}
		out = append(out, n)
	}
	return out
}

func TestFindGroupsIdenticalContent(t *testing.T) {
	dir := t.TempDir()
	big := pattern(100<<10, 'a')
	middleDiffers := slices.Clone(big)
	middleDiffers[50<<10] ^= 0xff

	write(t, dir, "a.flac", big)
	write(t, dir, "sub/b.flac", big)
	write(t, dir, "c.flac", middleDiffers)
	write(t, dir, "d.flac", big[:len(big)-1])
	write(t, dir, "s1.mp3", []byte("hello"))
	write(t, dir, "s2.MP3", []byte("hello"))
	write(t, dir, "s3.mp3", []byte("world"))
	write(t, dir, "e1.mp3", nil)
	write(t, dir, "e2.mp3", nil)
	write(t, dir, "notes.txt", []byte("hello"))

	r := find(t, dir, ".mp3", ".flac")
	want := [][]string{{"a.flac", "b.flac"}, {"s1.mp3", "s2.MP3"}}
	if got := names(r); !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if r.Scanned != 7 {
		t.Fatalf("scanned %d, want 7", r.Scanned)
	}
}

func TestFindCollapsesHardlinks(t *testing.T) {
	dir := t.TempDir()
	a := write(t, dir, "a.mp3", []byte("same"))
	if err := os.Link(a, filepath.Join(dir, "a-link.mp3")); err != nil {
		t.Skip("hardlinks unsupported:", err)
	}
	if got := names(find(t, dir)); len(got) != 0 {
		t.Fatalf("hardlinks reported as duplicates: %v", got)
	}
}

func TestVerifyDetectsChange(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.mp3", []byte("same"))
	write(t, dir, "b.mp3", []byte("same"))
	f := find(t, dir).Groups[0][1]

	if err := dedup.Verify(f); err != nil {
		t.Fatal(err)
	}
	later := f.ModTime.Add(time.Second)
	if err := os.Chtimes(f.Path, later, later); err != nil {
		t.Fatal(err)
	}
	if err := dedup.Verify(f); !errors.Is(err, dedup.ErrChanged) {
		t.Fatalf("got %v, want ErrChanged", err)
	}
}
