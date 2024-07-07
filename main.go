package main

import (
	"bytes"
	"fmt"
	"os"
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

	ch := make(chan []byte, 10)
	done := make(chan error)
	chanres := []byte{}

	go func() {
	L:
		for {
			select {
			case res := <-ch:
				fmt.Printf("Appending %d bytes...\r\n", len(res))
				chanres = append(chanres, res...)
			case don := <-done:
				if don != nil {
					panic(fmt.Errorf("an error occured while writing the data: %e", don))
				}
				break L
			default:
				continue
			}
		}
	}()

	err = pa.ToChannel(ch)
	if err != nil {
		panic(err)
	}
	// done <- nil

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
