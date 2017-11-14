/*
Package intern implements a string interner. This converts strings to integer indices, and also
converts the indices back to strings. The indices start at zero and increase by 1 for each new
unique string.

This can be used to reduce string duplication, and can also help by converting string keys into
more convenient integer IDs that are faster to compare. This can be particularly helpful when
using graph algorithms.

Intern is intended to be faster and perhaps more memory efficient than the obvious implementation
using a map[string]int32 and a []string. At the moment it is 'more than twice as quick' and uses
less than a 3rd of the RAM when saving UUIDs.
*/
package intern

import (
	"hash"
	"math/bits"
	"unsafe"

	"github.com/spaolacci/murmur3"
)

// IndexType is the type for the index we convert strings to. We use 32bits as surely no-one will have
// more than 4 billion strings?
type IndexType int32

// entry is an entry in our hash table. It stores the hash of the string that corresponds to this entry,
// and the index of the string in the string table. We actually store index+1 so we can use zero to
// indicate an empty entry
type entry struct {
	// We keep the hash alongside each entry to make it much faster to resize
	// It also speeds up stepping through entries when hashes clash
	hash uint32
	// Index is the index of the string in the strings table
	index IndexType
}

// Intern is a string-interner. It converts strings to integers and vice-versa. The
// integer indexes start at 0 and increase by 1 for each new string.
type Intern struct {
	clashes int
	entries []entry

	// Strings we want hash. In the end this will be external, but we'll need some kind of lookup function
	strings    []*[1024]string
	count      int
	threshold  int
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
		strings:    make([]*[1024]string, 0),
		loadFactor: loadFactor,
		threshold:  int(loadFactor * float64(cap)),
		hash:       murmur3.New32(),
	}
}

// Clashes returns the number of hash collisions we encounter adding strings.
func (i *Intern) Clashes() int { return i.clashes }

// Cap returns the number of strings that could be stored in the intern table
func (i *Intern) Cap() int { return len(i.entries) }

// Len returns the number of strings stored in the intern table
func (i *Intern) Len() int { return i.count }

// StringToIndex converts a string to an integer. The same string will always result in the same
// integer value. Values start at 0 and increment by one for each new unique string
func (i *Intern) StringToIndex(val string) IndexType {
	i.resize()
	// Hash the string
	hashVal := i.genhash(val)
	// Look up the string in the buckets
	entries := i.entries
	l := len(entries)
	cursor := int(hashVal) & (l - 1)
	start := cursor
	for entries[cursor].index != 0 {
		e := &entries[cursor]
		if e.hash == hashVal && i.IndexToString(e.index-1) == val {
			return e.index - 1
		}
		i.clashes++
		cursor++
		if cursor == l {
			cursor = 0
		}
		if cursor == start {
			panic("out of space!")
		}
	}

	// String was not found. Add the new string
	index := IndexType(i.count)
	i.count++
	j, k := index/1024, index&1023
	if k == 0 {
		i.strings = append(i.strings, new([1024]string))
	}
	i.strings[j][k] = val

	// Index starts at 0, but we use 0 to mean empty in the hash buckets
	entries[cursor] = entry{index: index + 1, hash: hashVal}
	return index
}

// IndexToString returns the string corresponding to the requested index.
func (i *Intern) IndexToString(index IndexType) string {
	return i.strings[index/1024][index&1023]
}

func (i *Intern) genhash(val string) uint32 {
	i.hash.Reset()
	i.hash.Write(*(*[]byte)(unsafe.Pointer(&val)))
	return i.hash.Sum32()
}

func (i *Intern) resize() {
	if i.count < i.threshold {
		return
	}

	// Make a new set of buckets twice as large as the current set
	oldEntries := i.entries
	numEntries := 2 * len(oldEntries)
	i.threshold = int(float64(numEntries) * i.loadFactor)
	i.entries = make([]entry, numEntries)
	mask := numEntries - 1

	for _, e := range oldEntries {
		if e.index == 0 {
			continue
		}
		cursor := int(e.hash) & mask
		for i.entries[cursor].index != 0 {
			cursor++
			if cursor == numEntries {
				cursor = 0
			}
		}
		i.entries[cursor] = e
	}
}
