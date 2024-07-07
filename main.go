package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/scheiblingco/go-pxar/pxar"
)

// The general process for the encoding is as follows:
// 1. Create a PBSArchive struct

// 2. Add files/folders. The stat/size information for each file/dir will be stored in the noderef structs.
//    The file data/content will be read just before encoding. All the information necessary to determine the positions
//    and lengths of all the data structs is in those structs, meaning we can take an async approach to reading
//    and encoding the archive since we know where everything needs to go in advance without reading all data into memory

// 3. Write the archive to a buffer or stream, and then to file

// 4. Create the catalog. Currently iterates over the tree again, should in the future be an option to generate this while writing the archive.

func main() {
	unordered := []pxar.GoodbyeItem{}
	// intsl := []int{}
	// tree := &BST{}

	for x := 0; x < 10; x++ {
		unordered = append(unordered, pxar.GoodbyeItem{
			Hash:   uint64(x),
			Offset: uint64(x),
			Length: uint64(x),
		})
		// tree.insert(x)
		// intsl = append(intsl, x)
	}
	// 6381579024

	binTree := make([]pxar.GoodbyeItem, len(unordered))
	pxar.GetBinaryHeap(unordered, &binTree)

	// tree := make([]pxar.GoodbyeItem, len(unordered))
	// tree2 := make([]pxar.GoodbyeItem, len(unordered))

	// pxar.MakeBinaryTree(unordered, &tree)
	// pxar.MakeBinaryTree(unordered, &tree2)

	// other := make([]pxar.GoodbyeItem, len(unordered))

	fmt.Println(unordered)

	// Create a new PXAR archive
	pa := PBSArchive{
		Filename: "test.pxar",
	}

	// Add one or multiple folders/files
	// If there isn't a single top-level directory, one will be automatically created
	// because the PXAR format requires it
	pa.AddFolder("./test-enc")

	// Create a buffer or stream to write the archive to
	buf := bytes.NewBuffer([]byte{})

	// Write the PXAR file to the buffer
	err := pa.ToBuffer(buf)
	if err != nil {
		panic(err)
	}

	// Create a PXAR file
	fa, err := os.OpenFile("demo.pxar", os.O_CREATE|os.O_WRONLY, 06444)
	if err != nil {
		panic(err)
	}
	defer fa.Close()

	// Write the buffer to the file
	_, err = fa.Write(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Write the catalogue/pxar file to the buffer
	catbuf := bytes.NewBuffer([]byte{})
	err = pa.WriteCatalogue(catbuf)
	if err != nil {
		panic(err)
	}

	// Create a catalogue file
	fc, err := os.OpenFile("demo.pcat1", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer fc.Close()

	// Write the catalogue buffer to the file
	_, err = fc.Write(catbuf.Bytes())
	if err != nil {
		panic(err)
	}
}
