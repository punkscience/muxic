// Package dedup finds files with identical content, reading as little as possible:
// size, then file identity, then the first and last 16 KiB, then the full content.
package dedup

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type File struct {
	Path    string
	Size    int64
	ModTime time.Time
	ID      ID
	info    fs.FileInfo
}

// Group holds files with identical content, sorted by path.
type Group []File

type Options struct {
	// Extensions limits the scan to these lower-case extensions, e.g. ".mp3". Empty means all files.
	Extensions []string
	Workers    int
	Warn       func(error)
}

type Result struct {
	Scanned int
	Groups  []Group
}

func Find(root string, opts Options) (Result, error) {
	if opts.Warn == nil {
		opts.Warn = func(error) {}
	}
	files, err := walk(root, opts)
	if err != nil {
		return Result{}, err
	}
	groups := bySize(files)
	groups = withIdentity(groups, opts)
	groups = regroup(groups, opts, func(Group) bool { return true }, edgeSum)
	groups = regroup(groups, opts, func(g Group) bool { return !coversWholeFile(g[0].Size) }, fullSum)
	sortGroups(groups)
	return Result{Scanned: len(files), Groups: groups}, nil
}

var ErrChanged = errors.New("changed since scan")

// Verify confirms f still matches what was scanned, guarding against edits between hashing and acting.
func Verify(f File) error {
	info, err := os.Lstat(f.Path)
	if err != nil {
		return err
	}
	id, err := identify(f.Path, info)
	if err != nil {
		return err
	}
	if info.Size() != f.Size || !info.ModTime().Equal(f.ModTime) || (f.ID.Known() && !id.SameObject(f.ID)) {
		return fmt.Errorf("%s: %w", f.Path, ErrChanged)
	}
	return nil
}

// Reclaims reports the bytes deleting f frees; a name sharing an inode with other names frees nothing.
func Reclaims(f File) int64 {
	if f.ID.Links > 1 {
		return 0
	}
	return f.Size
}

func walk(root string, opts Options) ([]File, error) {
	var files []File
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			opts.Warn(err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || !wanted(path, opts.Extensions) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			opts.Warn(fmt.Errorf("stat %s: %w", path, err))
			return nil
		}
		if info.Size() == 0 {
			return nil
		}
		files = append(files, File{Path: path, Size: info.Size(), ModTime: info.ModTime(), info: info})
		return nil
	})
	return files, err
}

func wanted(path string, exts []string) bool {
	return len(exts) == 0 || slices.Contains(exts, strings.ToLower(filepath.Ext(path)))
}

func bySize(files []File) []Group {
	m := map[int64]Group{}
	for _, f := range files {
		m[f.Size] = append(m[f.Size], f)
	}
	var groups []Group
	for _, g := range m {
		if len(g) > 1 {
			groups = append(groups, g)
		}
	}
	return groups
}

func withIdentity(groups []Group, opts Options) []Group {
	var out []Group
	for _, g := range groups {
		var kept Group
		for _, f := range g {
			id, err := identify(f.Path, f.info)
			if err != nil {
				opts.Warn(fmt.Errorf("identify %s: %w", f.Path, err))
				continue
			}
			f.ID = id
			if !slices.ContainsFunc(kept, func(k File) bool { return k.ID.SameObject(id) }) {
				kept = append(kept, f)
			}
		}
		if len(kept) > 1 {
			out = append(out, kept)
		}
	}
	return out
}

type job struct {
	group int
	file  File
}

type keyed struct {
	group int
	sum   digest
}

// regroup splits each selected group by the digest sum produces, in parallel, dropping singletons.
func regroup(groups []Group, opts Options, selected func(Group) bool, sum func(File) (digest, error)) []Group {
	var out []Group
	var jobs []job
	for i, g := range groups {
		if !selected(g) {
			out = append(out, g)
			continue
		}
		for _, f := range g {
			jobs = append(jobs, job{group: i, file: f})
		}
	}
	slices.SortFunc(jobs, func(a, b job) int { return cmp.Compare(a.file.ID.Ino, b.file.ID.Ino) })

	results := make(map[keyed]Group)
	var mu sync.Mutex
	queue := make(chan job)
	var wg sync.WaitGroup
	for range max(opts.Workers, 1) {
		wg.Go(func() {
			for j := range queue {
				d, err := sum(j.file)
				if err != nil {
					opts.Warn(fmt.Errorf("hash %s: %w", j.file.Path, err))
					continue
				}
				k := keyed{group: j.group, sum: d}
				mu.Lock()
				results[k] = append(results[k], j.file)
				mu.Unlock()
			}
		})
	}
	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	wg.Wait()

	for _, g := range results {
		if len(g) > 1 {
			out = append(out, g)
		}
	}
	return out
}

func sortGroups(groups []Group) {
	for _, g := range groups {
		slices.SortFunc(g, func(a, b File) int { return cmp.Compare(a.Path, b.Path) })
	}
	slices.SortFunc(groups, func(a, b Group) int {
		return cmp.Or(cmp.Compare(b[0].Size, a[0].Size), cmp.Compare(a[0].Path, b[0].Path))
	})
}
