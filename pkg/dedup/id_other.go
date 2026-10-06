//go:build !unix && !windows

package dedup

import "io/fs"

func identify(string, fs.FileInfo) (ID, error) {
	return ID{}, nil
}
