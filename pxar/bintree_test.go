package pxar_test

import (
	"testing"

	"github.com/scheiblingco/go-pxar/pxar"
)

// From https://github.com/proxmox/pxar/blob/aad6fb706ce4f0ee1fa211319226212ea08deeaf/src/binary_tree_array.rs#L196
func TestBinaryHeap(t *testing.T) {
	cases := map[int][]int{
		0:  {},
		1:  {0},
		2:  {1, 0},
		3:  {1, 0, 2},
		4:  {2, 1, 3, 0},
		5:  {3, 1, 4, 0, 2},
		6:  {3, 1, 5, 0, 2, 4},
		7:  {3, 1, 5, 0, 2, 4, 6},
		8:  {4, 2, 6, 1, 3, 5, 7, 0},
		9:  {5, 3, 7, 1, 4, 6, 8, 0, 2},
		10: {6, 3, 8, 1, 5, 7, 9, 0, 2, 4},
		11: {7, 3, 9, 1, 5, 8, 10, 0, 2, 4, 6},
		12: {7, 3, 10, 1, 5, 9, 11, 0, 2, 4, 6, 8},
		13: {7, 3, 11, 1, 5, 9, 12, 0, 2, 4, 6, 8, 10},
		14: {7, 3, 11, 1, 5, 9, 13, 0, 2, 4, 6, 8, 10, 12},
		15: {7, 3, 11, 1, 5, 9, 13, 0, 2, 4, 6, 8, 10, 12, 14},
		16: {8, 4, 12, 2, 6, 10, 14, 1, 3, 5, 7, 9, 11, 13, 15, 0},
		17: {9, 5, 13, 3, 7, 11, 15, 1, 4, 6, 8, 10, 12, 14, 16, 0, 2},
	}

	for n, want := range cases {
		list := make([]pxar.GoodbyeItem, n)
		for i := 0; i < n; i++ {
			list[i] = pxar.GoodbyeItem{
				Hash:   uint64(i),
				Offset: uint64(i),
				Length: uint64(i),
			}
		}

		tree := make([]pxar.GoodbyeItem, n)
		pxar.GetBinaryHeap(list, &tree)

		for i, v := range want {
			if int(tree[i].Hash) != v {
				t.Errorf("GetBinaryHeap(%d) = %d has %d, want %d", n, i, want[i], tree[i].Hash)
				break
			}
		}

	}
}
