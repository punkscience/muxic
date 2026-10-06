//go:build unix

package dedup

import (
	"io/fs"
	"syscall"
)

func identify(_ string, info fs.FileInfo) (ID, error) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ID{}, nil
	}
	return ID{Dev: uint64(st.Dev), Ino: uint64(st.Ino), Links: uint64(st.Nlink)}, nil
}
