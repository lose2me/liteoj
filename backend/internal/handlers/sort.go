package handlers

import "sort"

// sortSliceStable is a generic wrapper so callers don't have to write closure
// boilerplate. Not using slices.SortStableFunc to stay compatible with go 1.22
// without the experimental generic package.
func sortSliceStable[T any](s []T, less func(a, b T) bool) {
	sort.SliceStable(s, func(i, j int) bool { return less(s[i], s[j]) })
}
