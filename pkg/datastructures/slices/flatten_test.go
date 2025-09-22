package slices

import (
	"strings"
	"testing"
)

func Log2(n int) int {
	if n <= 0 {
		return -1
	}
	log := 0
	for n > 1 {
		n >>= 1
		log++
	}
	return log
}

func MakeIntSlice(n int) []int {
	s := make([]int, n)
	for i := range s {
		// Since all the quantities are nonnegative, we'll have the test values be exclusively negative.
		s[i] = -1 - i
	}
	return s
}

// Size represents the size of each individual sub-slice.
type Size int

const (
	One Size = iota
	Small
	Medium
	Large
	ExtraLarge
)

func (s Size) String() string {
	return [...]string{"One", "Small", "Medium", "Large", "ExtraLarge"}[s]
}
func (s Size) ToInt() int {
	return 1 << (s * 4)
}

// Quantity represents the number of sub-slices.
type Quantity int

const (
	Single Quantity = iota
	Few
	Several
	Many
	Lots
)

func (q Quantity) String() string {
	return [...]string{"Single", "Few", "Several", "Many", "Lots"}[q]
}
func (q Quantity) ToInt() int {
	return 1 << (q * 2)
}

// Order represents the order of sizes of the sub-slices.
// For a given Quantity, the Size represents an upper bound on the size of each sub-slice.
// If the order is not Equal, the the sizes will vary geometrically between 1 and Size,
// with the only difference being the order in which they appear.
type Order int

const (
	Equal Order = iota
	Ascending
	Descending
	Alternating
)

func (o Order) String() string {
	return [...]string{"Equal", "Ascending", "Descending", "Alternating"}[o]
}

func buildInput(order Order, size Size, quantity Quantity) [][]int {
	sizes := make([]int, quantity.ToInt())
	switch order {
	case Equal:
		for i := range sizes {
			sizes[i] = size.ToInt()
		}
	case Ascending:
		b := Log2(size.ToInt())
		for i := range sizes {
			sizes[i] = 1 << (b * (i + 1) / quantity.ToInt())
		}
	case Descending:
		b := Log2(size.ToInt())
		for i := range sizes {
			sizes[i] = 1 << (b * (i + 1) / quantity.ToInt())
		}
		copySizes := make([]int, len(sizes))
		copy(copySizes, sizes)
		for i := range sizes {
			sizes[i] = copySizes[len(copySizes)-1-i]
		}
	case Alternating:
		b := Log2(size.ToInt())
		for i := range sizes {
			sizes[i] = 1 << (b * (i + 1) / quantity.ToInt())
		}
		copySizes := make([]int, len(sizes))
		copy(copySizes, sizes)
		for i := range sizes {
			if i%2 == 0 {
				sizes[i] = copySizes[len(copySizes)-1-i]
			} else {
				sizes[i] = copySizes[i]
			}
		}
	}

	slices := make([][]int, quantity.ToInt())
	for i := range slices {
		slices[i] = MakeIntSlice(sizes[i])
	}

	return slices
}

func CompareWithFlattened(t *testing.T, want [][]int, got []int) {
	offset := 0
	for _, a := range want {
		for i, e := range a {
			if got[offset+i] != e {
				// End the test without showing all of the mismatches, which may be huge.
				t.Fatalf("mismatch at offset %d (sub-slice index %d, element index %d): got %d, want %d", offset+i, offset, i, got[offset+i], e)
			}
		}
		offset += len(a)
	}
}

func TestFlattenAppend(t *testing.T) {
	// We run through an abbreviated set of combinations here to keep the test time reasonable.
	for quantity := Single; quantity <= Several; quantity++ {
		for size := One; size <= Medium; size++ {
			for _, order := range []Order{Equal, Ascending, Alternating} {
				nameComponents := []string{
					quantity.String(),
					size.String(),
					order.String(),
				}
				t.Run(strings.Join(append(nameComponents, "Append"), "_"), func(t *testing.T) {
					want := buildInput(order, size, quantity)
					got := FlattenAppend(want)
					CompareWithFlattened(t, want, got)
				})
			}
		}
	}
}

func TestFlattenAllocate(t *testing.T) {
	// We run through an abbreviated set of combinations here to keep the test time reasonable.
	for quantity := Single; quantity <= Several; quantity++ {
		for size := One; size <= Medium; size++ {
			for _, order := range []Order{Equal, Ascending, Alternating} {
				nameComponents := []string{
					quantity.String(),
					size.String(),
					order.String(),
				}
				t.Run(strings.Join(append(nameComponents, "Allocate"), "_"), func(t *testing.T) {
					want := buildInput(order, size, quantity)
					got := FlattenAllocate(want)
					CompareWithFlattened(t, want, got)
				})
			}
		}
	}
}

// Note that these benchmarks run the exact same input through each function and each run.
func BenchmarkFlattenQuiet(b *testing.B) {
	for quantity := Single; quantity <= Lots; quantity++ {
		for size := One; size <= ExtraLarge; size++ {
			for order := Equal; order <= Alternating; order++ {
				nameComponents := []string{
					quantity.String(),
					size.String(),
					order.String(),
				}
				input := buildInput(order, size, quantity)
				b.Run(strings.Join(append(nameComponents, "Append"), "_"), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						FlattenAppend(input)
					}
				})
				b.Run(strings.Join(append(nameComponents, "Allocate"), "_"), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						FlattenAllocate(input)
					}
				})
			}
		}
	}
}
