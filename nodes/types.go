package nodes

import (
	"bytes"
	"encoding/binary"
	"math/bits"
	"os"
	"path/filepath"
	"syscall"

	"github.com/scheiblingco/go-pxar/pxar"
)

// A custom struct to hold only the stat/stat_t information that we need about the file
type Fstat struct {
	Mode       uint64
	Uid        uint32
	Gid        uint32
	Size       uint64
	MtimeSecs  uint64
	MtimeNsecs uint32
}

// All filesystem types need to implement the NodeRef interface
type NodeRef interface {
	// Should be empty NodeRef slice for all non-folder
	GetChildren() []NodeRef

	// Get a siphash of the filename
	GetHash() uint64

	// Write the payload of the node and any children to the buffer, return the written bytes and any errors
	WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error)

	// Write the catalogue of the node and any children to the buffer, return the written bytes and any errors
	// For directories, it will write it's own entry (including children) to the buffer AND return itself to the
	// parent directory for inclusion in the parent table.
	WriteCatalogue(buf *bytes.Buffer, pos *uint64, parentStartPos uint64) ([]byte, uint64, error)
}

// From https://github.com/protocolbuffers/protobuf-go/blob/v1.34.2/encoding/protowire/wire.go#L371
// SizeVarint returns the encoded size of a varint.
// The size is guaranteed to be within 1 and 10, inclusive.
func SizeUvarint(v uint64) int {
	// This computes 1 + (bits.Len64(v)-1)/7.
	// 9/64 is a good enough approximation of 1/7
	return int(9*uint32(bits.Len64(v))+64) / 64
}

func MakeUvarint(val uint64) []byte {
	buf := make([]byte, SizeUvarint(val))
	binary.PutUvarint(buf, val)
	return buf
}

type FolderRef struct {
	AbsPath string
	Name    string
	Stat    Fstat

	IsRoot          bool
	Children        []NodeRef
	GoodbyeItems    []pxar.GoodbyeItem
	CatalogueDirPos uint64
}

type FileRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type SymlinkRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type HardlinkRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type BlockDeviceRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type CharDeviceRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type FifoRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

type SocketRef struct {
	AbsPath string
	Name    string
	Stat    Fstat
}

func ReadNode(path string, isroot bool) NodeRef {
	info, err := os.Lstat(path)
	if err != nil {
		panic(err)
	}

	statT := info.Sys().(*syscall.Stat_t)
	fstat := Fstat{
		Mode:       uint64(statT.Mode),
		Uid:        statT.Uid,
		Gid:        statT.Gid,
		Size:       uint64(statT.Size),
		MtimeSecs:  uint64(statT.Mtim.Sec),
		MtimeNsecs: uint32(statT.Mtim.Nsec),
	}

	if info.Mode()&os.ModeSymlink != 0 {
		nref := &SymlinkRef{
			AbsPath: path,
			Name:    info.Name(),
			Stat:    fstat,
		}

		return nref
	}

	if info.IsDir() {
		nref := &FolderRef{
			AbsPath: path,
			Name:    info.Name(),
			Stat:    fstat,
		}

		files, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			nref.Children = append(nref.Children, ReadNode(filepath.Join(path, file.Name()), false))
		}

		return nref
	}

	if info.Mode().IsRegular() {
		nref := &FileRef{
			AbsPath: path,
			Name:    info.Name(),
			Stat:    fstat,
		}

		return nref
	}

	return nil
}
