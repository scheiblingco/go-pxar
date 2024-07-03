package pxar

import "math/bits"

func pow_of_2(e uint64) uint64 {
	return 1 << e
}

func log_of_2(k uint64) uint64 {
	return 8*8 - uint64(bits.LeadingZeros64(k)) - 1
}

func make_bst_inner(input []GoodbyeItem, n uint64, e uint64, output *[]GoodbyeItem, i uint64) {
	if n == 0 {
		return
	}
	p := pow_of_2(e - 1)
	q := pow_of_2(e)
	var k uint64
	if n >= p-1+p/2 {
		k = (q - 2) / 2
	} else {
		v := p - 1 + p/2 - n
		k = (q-2)/2 - v
	}

	(*output)[i] = input[k]

	make_bst_inner(input, k, e-1, output, i*2+1)
	make_bst_inner(input[k+1:], n-k-1, e-1, output, i*2+2)
}

func MakeBinaryTree(input []GoodbyeItem, output *[]GoodbyeItem) {
	n := uint64(len(input))
	make_bst_inner(input, n, log_of_2(n)+1, output, 0)
}
