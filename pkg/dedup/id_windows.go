//go:build windows

package dedup

import (
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

func identify(path string, _ fs.FileInfo) (ID, error) {
	f, err := os.Open(path)
	if err != nil {
		return ID{}, err
	}
	defer f.Close()

	var d windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &d); err != nil {
		return ID{}, err
	}
	return ID{
		Dev:   uint64(d.VolumeSerialNumber),
		Ino:   uint64(d.FileIndexHigh)<<32 | uint64(d.FileIndexLow),
		Links: uint64(d.NumberOfLinks),
	}, nil
}
