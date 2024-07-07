package main

import (
	"bytes"
	"fmt"
	"syscall"
	"time"

	"github.com/scheiblingco/go-pxar/nodes"
	"github.com/scheiblingco/go-pxar/pxar"
)

type PBSArchiveInterface interface {
	// Add Folder
	AddFolder(path string)

	// Add File
	AddFile(path string)

	// Write to buffer
	ToBuffer(buf *bytes.Buffer) error

	// Create catalogue
	WriteCatalogue(buf *bytes.Buffer) error
}

type PBSArchive struct {
	// Directory Tree
	Trees []nodes.NodeRef

	// Filename of the resulting archive
	Filename string
}

// Add a top-level folder to the archive
func (pa *PBSArchive) AddFolder(path string) {
	pa.Trees = append(pa.Trees, nodes.ReadNode(path, true, ""))
}

// Add a "top-level" file to the archive
func (pa *PBSArchive) AddFile(path string) {
	pa.Trees = append(pa.Trees, nodes.ReadNode(path, true, ""))
}

// Returns a single node that we can use as a top-level node in the archive.
// If the archive consists of a single folder, that folder is used as the top-level node.
// If there are multiple folders, a single file or a mix of files and folders at the top
// a virtual top-level dir will be created (owner: 0, group: 0, mode: 0777)
func (pa *PBSArchive) GetParentNode() (*nodes.FolderRef, error) {
	var topTree *nodes.FolderRef

	if _, ok := pa.Trees[0].(*nodes.FolderRef); ok && len(pa.Trees) == 1 {
		if len(pa.Trees[0].GetChildren()) == 0 {
			return nil, fmt.Errorf("only blank directory, no files to backup")
		}

		if _, ok := pa.Trees[0].(*nodes.FolderRef); !ok {
			return nil, fmt.Errorf("top level item must be a directory. add multiple items/files/folders to create a virtual top directory")
		}

		topTree = pa.Trees[0].(*nodes.FolderRef)
	} else {
		// Create a virtual topdir if only files or multiple root folders are being backed up
		for ti := range pa.Trees {
			if _, ok := pa.Trees[ti].(*nodes.FolderRef); ok {
				pa.Trees[ti].(*nodes.FolderRef).IsRoot = false
			}
		}

		// The virtual top-level directory
		topTree = &nodes.FolderRef{
			IsRoot:  true,
			AbsPath: "/",
			Name:    pa.Filename,
			Stat: nodes.Fstat{
				Mode:       syscall.S_IFDIR | 0777,
				Uid:        0,
				Gid:        0,
				Size:       0,
				MtimeSecs:  uint64(time.Now().Unix()),
				MtimeNsecs: uint32(time.Now().UnixNano()),
			},
			Children: pa.Trees,
		}
	}

	return topTree, nil
}

// Writes the pxar archive to a buffer.
func (pa *PBSArchive) ToBuffer(buf *bytes.Buffer) error {
	if len(pa.Trees) == 0 {
		return fmt.Errorf("no items to write")
	}

	// Get the parent node, and call the recursive WritePayload function
	// to write all data to the buffer
	pos := uint64(0)
	topTree, err := pa.GetParentNode()
	if err != nil {
		return err
	}

	topTree.IsRoot = true
	topTree.Name = pa.Filename + ".didx"

	_, err = topTree.WritePayload(buf, &pos)
	if err != nil {
		return err
	}

	fmt.Printf("Write buffer finished on pos %d with len %d\r\n", pos, buf.Len())

	return nil
}

// Writes the pxar archive to a buffer.
func (pa *PBSArchive) ToChannel(ch chan []byte) error {
	if len(pa.Trees) == 0 {
		return fmt.Errorf("no items to write")
	}

	// Get the parent node, and call the recursive WritePayload function
	// to write all data to the buffer
	pos := uint64(0)
	topTree, err := pa.GetParentNode()
	if err != nil {
		return err
	}

	topTree.IsRoot = true
	topTree.Name = pa.Filename + ".didx"

	_, err = topTree.WritePayloadChannel(ch, &pos)
	if err != nil {
		return err
	}

	fmt.Printf("Write buffer finished on pos %d\r\n", pos)

	return nil
}

// Writes the catalogue to a buffer. The catalogue is a list of all the nodes in the archive
// together with references to their parent nodes. The catalogue is used to quickly locate
// a node in the archive by its path.
func (pa *PBSArchive) WriteCatalogue(buf *bytes.Buffer) error {
	// Write magic header
	buf.Write(pxar.CatalogMagic)

	// Get parent
	pos := uint64(0)
	topTree, err := pa.GetParentNode()
	if err != nil {
		return err
	}

	topTree.IsRoot = true
	topTree.Name = pa.Filename + ".didx"

	lastBytes, n, err := topTree.WriteCatalogue(buf, &pos, pos)
	if err != nil {
		return err
	}

	bufLen := uint64(buf.Len())

	buf.Write(nodes.MakeUvarint(uint64(len(lastBytes) + 1)))
	buf.WriteByte(byte(0x01))
	buf.Write(lastBytes)

	buf.Write(nodes.MakeUvarint(uint64(bufLen)))
	for buf.Len()%8 != 0 {
		buf.WriteByte(0x00)
	}

	fmt.Printf("wrote %d bytes to catalogue\r\n", n)

	return nil
}
