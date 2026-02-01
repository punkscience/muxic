package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileEntry represents the cached metadata for a file.
type FileEntry struct {
	Signature string `json:"signature"`
	ModTime   int64  `json:"mod_time"`
	Size      int64  `json:"size"`
}

// Cache represents the mapping of file paths to their signatures.
type Cache map[string]FileEntry

// LoadCache loads the cache from the specified file.
// If the file does not exist, it returns an empty cache.
func LoadCache(path string) (Cache, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return make(Cache), nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cache Cache
	if err := json.NewDecoder(f).Decode(&cache); err != nil {
		// If decoding fails (e.g., empty or corrupt file), return empty cache
		return make(Cache), nil
	}
	return cache, nil
}

// SaveCache saves the cache to the specified file.
func SaveCache(path string, cache Cache) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cache)
}

// GenerateSignature computes the SHA-256 hash of the file content.
func GenerateSignature(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// UpdateEntry updates the cache entry for a file if necessary.
// It returns the signature and a boolean indicating if the signature was computed (fresh).
func UpdateEntry(path string, info os.FileInfo, cache Cache, mu *sync.Mutex) (string, bool, error) {
	// Check if entry exists and is up to date
	if mu != nil {
		mu.Lock()
	}
	entry, exists := cache[path]
	if mu != nil {
		mu.Unlock()
	}

	if exists && entry.ModTime == info.ModTime().Unix() && entry.Size == info.Size() {
		return entry.Signature, false, nil
	}

	// Compute new signature
	sig, err := GenerateSignature(path)
	if err != nil {
		return "", false, err
	}

	newEntry := FileEntry{
		Signature: sig,
		ModTime:   info.ModTime().Unix(),
		Size:      info.Size(),
	}

	if mu != nil {
		mu.Lock()
		cache[path] = newEntry
		mu.Unlock()
	} else {
		cache[path] = newEntry
	}

	return sig, true, nil
}
