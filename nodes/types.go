package nodes

import (
	"bytes"
	"encoding/binary"
	"math/bits"

	"github.com/scheiblingco/go-pxar/pxar"
)

type Fstat struct {
	Mode       uint64
	Uid        uint32
	Gid        uint32
	Size       uint64
	MtimeSecs  uint64
	MtimeNsecs uint32
}

type NodeRef interface {
	GetChildren() []NodeRef
	GetHash() uint64

	WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error)
	// WriteCatalogue(buf *bytes.Buffer, pos *uint64) (uint64, error)
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

	IsRoot       bool
	Children     []NodeRef
	GoodbyeItems []pxar.GoodbyeItem
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
