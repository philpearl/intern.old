package intern_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pborman/uuid"
	"github.com/philpearl/intern"
	"github.com/stretchr/testify/assert"
)

func TestSame(t *testing.T) {
	in := intern.New(5, 0.7)
	// Tests that the first index is zero and that the indexes then increase
	// by one
	for i := 0; i < 200; i++ {
		assert.EqualValues(t, i, in.StringToIndex(strconv.Itoa(i)))
	}
	// Tests we get the same answers if we ask again
	for i := 0; i < 200; i++ {
		assert.EqualValues(t, i, in.StringToIndex(strconv.Itoa(i)))
	}
}

func TestDifferent(t *testing.T) {
	in := intern.New(100, 0.7)
	assert.NotEqual(t, in.StringToIndex("hat"), in.StringToIndex("coat"))
	assert.Equal(t, 0, in.Clashes())
}

func BenchmarkIntern(b *testing.B) {
	const loadFactor = 0.7
	for _, loadFactor := range []float64{0.5, 0.6, 0.7, 0.8} {
		b.Run(fmt.Sprintf("loadfactor=%.2f", loadFactor), func(b *testing.B) {
			in := intern.New(b.N, 0.7)
			strings := make([]string, b.N)
			for i := range strings {
				strings[i] = uuid.New()
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				in.StringToIndex(strings[i])
			}
			b.Logf("%.2f%% clashes", 100*float64(in.Clashes())/float64(b.N))
		})
	}
}
