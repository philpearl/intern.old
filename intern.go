package intern

import (
	"hash"
	"math/bits"
	"unsafe"

	"github.com/spaolacci/murmur3"
)

type IndexType int32

type entry struct {
	// We keep the hash alongside each entry to make it much faster to resize
	hash uint32
	// Index is the index of the string in the strings table
	index IndexType
}

type Intern struct {
	clashes int
	entries []entry

	// Strings we want hash. In the end this will be external, but we'll need some kind of lookup function
	strings    []string
	loadFactor float64
	hash       hash.Hash32
}

func New(cap int, loadFactor float64) *Intern {
	if cap < 16 {
		cap = 16
	} else {
		cap = 1 << uint(64-bits.LeadingZeros(uint(cap-1)))
	}
	return &Intern{
		entries:    make([]entry, cap),
		strings:    make([]string, 0, cap),
		loadFactor: loadFactor,
		hash:       murmur3.New32(),
	}
}

func (i *Intern) Clashes() int { return i.clashes }

// Cap returns the number of strings that could be stored in the intern table
func (i *Intern) Cap() int { return len(i.entries) }

// Len returns the number of strings stored in the intern table
func (i *Intern) Len() int { return len(i.strings) }

// StringToIndex converts a string to an integer. The same string will always result in the same
// integer value. Values start at 0 and increment by one for each new unique string
func (i *Intern) StringToIndex(val string) IndexType {
	i.resize()
	// Hash the string
	hashVal := i.genhash(val)
	// Look up the string in the buckets
	cursor := int(hashVal) & (len(i.entries) - 1)
	start := cursor
	for {
		e := i.entries[cursor]
		if e.index == 0 {
			// This bucket is empty - val is not found
			break
		}
		if e.hash == hashVal && i.IndexToString(e.index-1) == val {
			return e.index - 1
		}
		i.clashes++
		cursor++
		if cursor == len(i.entries) {
			cursor = 0
		}
		if cursor == start {
			panic("out of space!")
		}
	}

	// String was not found. Add the new string
	i.strings = append(i.strings, val)
	index := IndexType(len(i.strings))
	i.entries[cursor] = entry{index: index, hash: hashVal}
	// Index starts at 0, but we use 0 to mean empty in the hash buckets
	return index - 1
}

// IndexToString returns the string corresponding to the requested index.
func (i *Intern) IndexToString(index IndexType) string {
	return i.strings[index]
}

func (i *Intern) genhash(val string) uint32 {
	i.hash.Reset()
	i.hash.Write(*(*[]byte)(unsafe.Pointer(&val)))
	return i.hash.Sum32()
}

func (i *Intern) resize() {
	if len(i.strings) < int(i.loadFactor*float64(len(i.entries))) {
		return
	}

	// Make a new set of buckets twice as large as the current set
	oldEntries := i.entries
	numEntries := 2 * len(oldEntries)
	i.entries = make([]entry, numEntries)

	for _, e := range oldEntries {
		if e.index == 0 {
			continue
		}
		cursor := int(e.hash) & (numEntries - 1)
		for i.entries[cursor].index != 0 {
			cursor++
			if cursor == numEntries {
				cursor = 0
			}
		}
		i.entries[cursor] = e
	}
}
