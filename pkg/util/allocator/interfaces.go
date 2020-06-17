package allocator

// Interface manages the allocation of items out of a range. Interface
// should be threadsafe.
type Interface interface {
	Allocate(int) (bool, error)
	AllocateNext() (int, bool, error)
	Release(int) error
	ForEach(func(int))

	// For testing
	Has(int) bool

	// For testing
	Free() int
}

// Snapshottable is an Interface that can be snapshotted and restored. Snapshottable
// should be threadsafe.
type Snapshottable interface {
	Interface
	Snapshot() (string, []byte)
	Restore(string, []byte) error
}

type Factory func(max int, rangeSpec string) Interface
