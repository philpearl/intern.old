package intern_test

import (
	"strconv"
	"testing"

	"github.com/pborman/uuid"
	"github.com/philpearl/intern"
)

// compare is the simple way to implement Intern with normal Go things
type compare struct {
	m map[string]intern.IndexType
	s []string
}

func newCompare(cap int) *compare {
	return &compare{
		m: make(map[string]intern.IndexType, cap),
		s: make([]string, 0, cap),
	}
}

func (c *compare) StringToIndex(val string) intern.IndexType {
	index, ok := c.m[val]
	if ok {
		return index - 1
	}
	c.s = append(c.s, val)
	index = intern.IndexType(len(c.s))
	c.m[val] = index
	return index - 1
}

func (c *compare) IndexToString(index intern.IndexType) string {
	return c.s[index-1]
}

func BenchmarkCompare(b *testing.B) {
	for _, numStrings := range []int{1000, 10000, 100000, 1000000} {
		b.Run(strconv.Itoa(numStrings), func(b *testing.B) {
			strings := make([]string, numStrings)
			for i := range strings {
				strings[i] = uuid.New()
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i += numStrings {
				c := newCompare(128)
				for _, s := range strings {
					c.StringToIndex(s)
				}
			}
		})
	}
}

func BenchmarkCompareIntern(b *testing.B) {
	for _, numStrings := range []int{1000, 10000, 100000, 1000000} {
		b.Run(strconv.Itoa(numStrings), func(b *testing.B) {
			strings := make([]string, numStrings)
			for i := range strings {
				strings[i] = uuid.New()
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i += numStrings {
				c := intern.New(128, 0.7)
				for _, s := range strings {
					c.StringToIndex(s)
				}
			}
		})
	}
}
