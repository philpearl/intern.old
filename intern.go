package intern

import (
	"unsafe"
)

type IndexType int32

type bucket struct {
	index IndexType
}

type Intern struct {
	clashes int
	buckets []bucket

	// Strings we want hash. In the end this will be external, but we'll need some kind of lookup function
	strings    []string
	loadFactor float64
}

func New(cap int, loadFactor float64) *Intern {
	if cap == 0 {
		cap = 10
	}
	return &Intern{
		buckets:    make([]bucket, cap),
		strings:    make([]string, 0, cap),
		loadFactor: loadFactor,
	}
}

func (i *Intern) Clashes() int { return i.clashes }

// Cap returns the number of strings that could be stored in the intern table
func (i *Intern) Cap() int { return len(i.buckets) }

// Len returns the number of strings stored in the intern table
func (i *Intern) Len() int { return len(i.strings) }

// StringToIndex converts a string to an integer. The same string will always result in the same
// integer value. Values start at 0 and increment by one for each new unique string
func (i *Intern) StringToIndex(val string) IndexType {
	i.resize()
	// Hash the string
	hashVal := i.genhash(val)
	// Look up the string in the buckets
	cursor := hashVal % len(i.buckets)
	start := cursor
	for {
		index := i.buckets[cursor].index
		if index == 0 {
			// This bucket is empty - val is not found
			break
		}
		if i.IndexToString(index-1) == val {
			return index - 1
		}
		i.clashes++
		cursor++
		if cursor == len(i.buckets) {
			cursor = 0
		}
		if cursor == start {
			panic("out of space!")
		}
	}

	// String was not found. Add the new string
	i.strings = append(i.strings, val)
	index := IndexType(len(i.strings))
	i.buckets[cursor].index = index
	// Index starts at 0, but we use 0 to mean empty in the hash buckets
	return index - 1
}

// IndexToString returns the string corresponding to the requested index.
func (i *Intern) IndexToString(index IndexType) string {
	return i.strings[index]
}

func (i *Intern) genhash(val string) int {
	var hash fnvHash = offset32
	hash.Write(*(*[]byte)(unsafe.Pointer(&val)))
	return int(hash)
}

func (i *Intern) resize() {
	if len(i.strings) < int(i.loadFactor*float64(len(i.buckets))) {
		return
	}

	// Make a new set of buckets twice as large as the current set
	oldBuckets := i.buckets
	numBuckets := 2 * len(oldBuckets)
	i.buckets = make([]bucket, numBuckets)

	for _, b := range oldBuckets {
		if b.index == 0 {
			continue
		}
		val := i.strings[b.index-1]
		hashVal := i.genhash(val)
		cursor := hashVal % numBuckets
		for i.buckets[cursor].index != 0 {
			cursor++
			if cursor == numBuckets {
				cursor = 0
			}
		}
		i.buckets[cursor].index = b.index
	}
}

// Taken from the hash/fnv package to speed it up a little
type fnvHash uint32

const (
	offset32 = 2166136261
	prime32  = 16777619
)

func (s *fnvHash) Write(data []byte) {
	hash := *s
	for _, c := range data {
		hash ^= fnvHash(c)
		hash *= prime32
	}
	*s = hash
}
