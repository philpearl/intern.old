
[![Build Status](https://travis-ci.org/philpearl/intern.svg)](https://travis-ci.org/philpearl/intern) [![GoDoc](https://godoc.org/github.com/philpearl/intern?status.svg)](https://godoc.org/github.com/philpearl/intern)


Intern is a string interning library. It creates a unique index for each string, and allows you to lookup the string for the index. The usual reason for this is to avoid duplicating strings.

Another use is to swap string keys for integer indices. This can be very useful when running certain algorithms. For example graph searches: in this case we need to compare keys a large number of times.

The indices start at 0 and increment by 1 for each new string, so can be used as indices into a slice of entries corresponding to each key.

```go
import "github.com/philpearl/intern"

func DoSomething() {
	in := intern.New(1000, 0.7) // 0.7 seems to be a good load factor

	alice := in.StringToIndex("Alice")
	bob := in.StringToIndex("Bob")

	fmt.Printf("Alice's name is %s, Bob's is %s", in.IndexToString(alice), in.IndexToString(bob))
}
```