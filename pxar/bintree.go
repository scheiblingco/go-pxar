package pxar

import (
	"math/bits"
	"sort"
)

func InsertBst(origin []GoodbyeItem, tree *[]GoodbyeItem, n, e, i uint64) {
	if n == 0 {
		return
	}

	p := uint64(1 << (e - 1))
	k := uint64((1<<e)/2 - 1)

	c := (3*p)/2 - 1

	if n < c {
		k = k + n - c
	}

	(*tree)[i] = origin[k]

	InsertBst(origin, tree, k, e-1, i*2+1)
	InsertBst(origin[k+1:], tree, n-k-1, e-1, i*2+2)
}

/* Generate a binary search tree stored in an array from a sorted array. Specifically, for any given sorted
* array 'input' of 'n' elements of size 'size' permute the array so that the following rule holds:
*
* For each array item with index i, the item at 2*i+1 is smaller and the item 2*i+2 is larger.
*
* This structure permits efficient (meaning: O(log(n)) binary searches: start with item i=0 (i.e. the root of
* the BST), compare the value with the searched item, if smaller proceed at item i*2+1, if larger proceed at
* item i*2+2, and repeat, until either the item is found, or the indexes grow beyond the array size, which
* means the entry does not exist. Effectively this implements bisection, but instead of jumping around wildly
* in the array during a single search we only search with strictly monotonically increasing indexes.
* Permutation formula originally by L. Bressel, 2017
 */

func GetBinaryHeap(origin []GoodbyeItem, tree *[]GoodbyeItem) {
	sort.Slice(origin, func(i, j int) bool {
		return origin[i].Hash < origin[j].Hash
	})

	originLen := uint64(len(origin))

	InsertBst(origin, tree, originLen, uint64(bits.Len64(originLen)), 0)
}
