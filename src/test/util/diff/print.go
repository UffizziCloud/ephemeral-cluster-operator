package diff

import "fmt"

// PrintWantGot takes a diff string generated by cmp.Diff and returns it
// in a consistent format for reuse across all of our tests. This
// func assumes that the order of arguments passed to cmp.Diff was
// (want, got) or, in other words, the expectedResult then the actualResult.
func PrintWantGot(diff string) string {
	return fmt.Sprintf("(-want, +got): %s", diff)
}