package dedup

// ID identifies the underlying file object, so several names for one inode compare equal.
// A zero ID means the platform could not tell.
type ID struct {
	Dev   uint64
	Ino   uint64
	Links uint64
}

func (id ID) Known() bool {
	return id.Dev != 0 || id.Ino != 0
}

func (id ID) SameObject(other ID) bool {
	return id.Known() && id.Dev == other.Dev && id.Ino == other.Ino
}
