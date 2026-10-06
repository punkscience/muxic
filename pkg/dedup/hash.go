package dedup

import (
	"crypto/sha256"
	"io"
	"os"
	"sync"
)

const edgeSize = 16 << 10

type digest [sha256.Size]byte

var bufPool = sync.Pool{New: func() any { b := make([]byte, 1<<20); return &b }}

func coversWholeFile(size int64) bool {
	return size <= 2*edgeSize
}

// edgeSum hashes the first and last edgeSize bytes, or the whole file when it is small.
func edgeSum(f File) (digest, error) {
	if coversWholeFile(f.Size) {
		return fullSum(f)
	}
	file, err := os.Open(f.Path)
	if err != nil {
		return digest{}, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.CopyN(h, file, edgeSize); err != nil {
		return digest{}, err
	}
	if _, err := file.Seek(-edgeSize, io.SeekEnd); err != nil {
		return digest{}, err
	}
	if _, err := io.CopyN(h, file, edgeSize); err != nil {
		return digest{}, err
	}
	return digest(h.Sum(nil)), nil
}

func fullSum(f File) (digest, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return digest{}, err
	}
	defer file.Close()

	buf := bufPool.Get().(*[]byte)
	defer bufPool.Put(buf)

	h := sha256.New()
	if _, err := io.CopyBuffer(h, struct{ io.Reader }{file}, *buf); err != nil {
		return digest{}, err
	}
	return digest(h.Sum(nil)), nil
}
