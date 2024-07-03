package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/scheiblingco/go-pxar/nodes"
)

type VirtualFileinfo struct {
	NameVal    string
	SizeVal    int64
	ModeVal    fs.FileMode
	ModTimeVal time.Time
	IsDirVal   bool
	SysVal     interface{}
}

func (vfi *VirtualFileinfo) Name() string {
	return vfi.NameVal
}
func (vfi *VirtualFileinfo) Size() int64 {
	return vfi.SizeVal
}
func (vfi *VirtualFileinfo) Mode() fs.FileMode {
	return vfi.ModeVal
}
func (vfi *VirtualFileinfo) ModTime() time.Time {
	return vfi.ModTimeVal
}
func (vfi *VirtualFileinfo) IsDir() bool {
	return vfi.IsDirVal
}
func (vfi *VirtualFileinfo) Sys() interface{} {
	return vfi.SysVal
}

func ReadNode(path string, isroot bool) nodes.NodeRef {
	info, err := os.Lstat(path)
	if err != nil {
		panic(err)
	}

	statT := info.Sys().(*syscall.Stat_t)
	fstat := nodes.Fstat{
		Mode:       uint64(statT.Mode),
		Uid:        statT.Uid,
		Gid:        statT.Gid,
		Size:       uint64(statT.Size),
		MtimeSecs:  uint64(statT.Mtim.Sec),
		MtimeNsecs: uint32(statT.Mtim.Nsec),
	}

	// if info.Mode()&os.ModeSymlink != 0 {
	// 	nref := &nodes.SymlinkRef{
	// 		AbsPath:  path,
	// 		Name: info.Name(),
	// 		Stat: fstat
	// 	}

	// 	return nref
	// }

	if info.IsDir() {
		nref := &nodes.FolderRef{
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
		nref := &nodes.FileRef{
			AbsPath: path,
			Name:    info.Name(),
			Stat:    fstat,
		}

		return nref
	}

	return nil
}

type PBSArchiveInterface interface {
	// Add Folder
	AddFolder(path string)

	// Add File
	AddFile(path string)

	// Write to file
	ToFile(path string) error

	// Write to buffer
	ToBuffer(buf *bytes.Buffer) error

	// Write to stream
	ToStream(stream *os.File) error

	// Create catalogue
	WriteCatalogue(buf *bytes.Buffer) error
}

type PBSArchive struct {
	// Directory Tree
	Trees []nodes.NodeRef

	// Filename
	Filename string
}

func (pa *PBSArchive) AddFolder(path string) {
	pa.Trees = append(pa.Trees, ReadNode(path, true))
}

func (pa *PBSArchive) AddFile(path string) {
	pa.Trees = append(pa.Trees, ReadNode(path, true))
}

func (pa *PBSArchive) ToFile(path string) error {
	fstr := fmt.Sprintf("%s.pxar", path)
	f, err := os.OpenFile(fstr, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bytes.NewBuffer([]byte{})
	err = pa.ToBuffer(buf)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d bytes to file\r\n", n)

	return nil
}

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

func (pa *PBSArchive) ToBuffer(buf *bytes.Buffer) error {
	if len(pa.Trees) == 0 {
		return fmt.Errorf("no items to write")
	}

	pos := uint64(0)
	topTree, err := pa.GetParentNode()
	if err != nil {
		return err
	}

	topTree.IsRoot = true

	_, err = topTree.WritePayload(buf, &pos)
	if err != nil {
		return err
	}

	// goodbyeRecord := []nodes.GoodbyeItem{}

	// for _, child := range topTree.GetChildren() {
	// 	child.WriteGoodbyeItem(&goodbyeRecord)
	// }

	// goodbyeTable, err := nodes.FinishGoodbyeTable(goodbyeRecord, &pos, &dirStart)
	// if err != nil {
	// 	return err
	// }

	// gbiDescriptor := pxar.PxarDescriptor{
	// 	EntryType: pxar.PXAR_GOODBYE,
	// 	Length:    uint64(16) + uint64(len(goodbyeTable)),
	// }

	// err = binary.Write(buf, binary.LittleEndian, gbiDescriptor)
	// if err != nil {
	// 	return err
	// }

	// pos += gbiDescriptor.Length

	// _, err = buf.Write(goodbyeTable)
	// if err != nil {
	// 	return err
	// }

	fmt.Printf("Write buffer finished on pos %d with len %d\r\n", pos, buf.Len())

	return nil
}

func main() {
	pa := PBSArchive{
		Filename: "test.pxar",
	}

	pa.AddFolder("/home/larsec/pxar-demo/test-enc")

	buf := bytes.NewBuffer([]byte{})
	pa.ToBuffer(buf)

	f, err := os.OpenFile("/home/larsec/pxar-demo/test-newgo.pxar", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	n, err := f.Write(buf.Bytes())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote %d bytes to file\r\n", n)

	fmt.Println("Hold")
}
