package pxar

import (
	"bytes"
	"io"
)

type CatalogEntryType uint8

const (
	PXAR_ENTRY               uint64 = 0xd5956474e588acef
	PXAR_ENTRY_V1            uint64 = 0x11da850a1c1cceff
	PXAR_FILENAME            uint64 = 0x16701121063917b3
	PXAR_SYMLINK             uint64 = 0x27f971e7dbf5dc5f
	PXAR_DEVICE              uint64 = 0x9fc9e906586d5ce9
	PXAR_XATTR               uint64 = 0x0dab0229b57dcd03
	PXAR_ACL_USER            uint64 = 0x2ce8540a457d55b8
	PXAR_ACL_GROUP           uint64 = 0x136e3eceb04c03ab
	PXAR_ACL_GROUP_OBJ       uint64 = 0x10868031e9582876
	PXAR_ACL_DEFAULT         uint64 = 0xbbbb13415a6896f5
	PXAR_ACL_DEFAULT_USER    uint64 = 0xc89357b40532cd1f
	PXAR_ACL_DEFAULT_GROUP   uint64 = 0xf90a8a5816038ffe
	PXAR_FCAPS               uint64 = 0x2da9dd9db5f7fb67
	PXAR_QUOTA_PROJID        uint64 = 0xe07540e82f7d1cbb
	PXAR_HARDLINK            uint64 = 0x51269c8422bd7275
	PXAR_PAYLOAD             uint64 = 0x28147a1b0b7c1a25
	PXAR_GOODBYE             uint64 = 0x2fec4fa642d5731d
	PXAR_GOODBYE_TAIL_MARKER uint64 = 0xef5eed5b753e1555

	IF_MT           uint64 = 0o0170000
	IF_SOCKET       uint64 = 0o0140000
	IF_LINK         uint64 = 0o0120000
	IF_REGULAR_FILE uint64 = 0o0100000
	IF_BLOCKDEV     uint64 = 0o0060000
	IF_DIR          uint64 = 0o0040000
	IF_CHR          uint64 = 0o0020000
	IF_IFO          uint64 = 0o0010000

	IS_UID uint64 = 0o0004000
	IS_GID uint64 = 0o0002000
	IS_VTX uint64 = 0o0001000

	// generated with:
	// $ echo -n 'PROXMOX ARCHIVE FORMAT' | sha1sum | sed -re 's/^(.{16})(.{16}).*$/0x\1, 0x\2/'
	PXAR_HASH_KEY_1 uint64 = 0x83ac3f1cfbb450db
	PXAR_HASH_KEY_2 uint64 = 0xaa4f1b6879369fbd

	DirectoryEntry   CatalogEntryType = 'd'
	FileEntry        CatalogEntryType = 'f'
	SymlinkEntry     CatalogEntryType = 'l'
	HardlinkEntry    CatalogEntryType = 'h'
	BlockDeviceEntry CatalogEntryType = 'b'
	CharDeviceEntry  CatalogEntryType = 'c'
	FifoEntry        CatalogEntryType = 'p'
	SocketEntry      CatalogEntryType = 's'
)

var (
	CatalogMagic = []byte{145, 253, 96, 249, 196, 103, 88, 213}
)

type PxarBlock interface {
	Write(buf *bytes.Buffer) error
	WriteStream(stream io.Writer) error
	WriteChannel(ch chan []byte) error
}

type PxarSection interface {
	Write(buf *bytes.Buffer, pos *uint64) (uint64, error)
	WriteStream(stream io.Writer, pos *uint64) (uint64, error)
	WriteChannel(ch chan []byte, pos *uint64) (uint64, error)
}

type PxarDescriptor struct {
	EntryType uint64
	Length    uint64
}
