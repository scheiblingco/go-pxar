# GO-PXAR
This is a go implementation** of the proxmox backup client. 

Please note that everything is still under development, and should not yet be trusted in production. 

I wanna begin by giving a special thanks to [https://github.com/tizbac/proxmoxbackupclient_go](Tiziano (tizbac)), who made a great deal of the hard work to figure out how the proxmox archive format (PXAR) together with the catalog format (pcat1) works despite the lack of documentation. If you need a working client that's been tested for more than 5 minutes, have a look at his implementation.

** This is not a full implementation.. yet.

## Current Functionality
- Create PXAR archives to disk
  - Support for single-file, multi-file, multi-rootdir archives (via a virtual top-level directory in the archive)
  - Support for files, directories, and symlinks
  - Verified 1:1 against the official pxar cli-client

- Create catalog (pcat1) files to disk
  - Support for files, directories, and symlinks
  - Verified against the official proxmox implementation in PBS-client

## Roadmap
- [ ] Misc. Todo
  - [X] Convert all paths to abspath in noderef instead of adding relative paths as-is, that way we can determine if a symlink target is in-tree as well.
  - [ ] Verify that link targets are in rootpath, possibly modify to convert to relative paths so they work no matter where an archive is unpacked
  - [ ] Concurrent Uploads/PXAR Encoding
  - [ ] Add a `--debug` flag to the CLI to enable debug logging
- [ ] PXAR Archives
  - [ ] PXAR Creation
    - [x] Files
    - [x] Directories
    - [x] Symlinks
    - [ ] Hardlinks
    - [ ] Devices/FIFO
    - [ ] Sockets
    - [ ] Extended Attributes
    - [ ] Manually Create PXAR from data in Go instead of from FS
  - [ ] PXAR Parsing/Unpacking
- [ ] Catalogs
  - [ ] Catalog Creation
    - [x] Files
    - [x] Directories
    - [x] Symlinks
    - [ ] Hardlinks
    - [ ] Devices/FIFO
    - [ ] Sockets
- [ ] Communication
  - [ ] PBS Client
  - [ ] FIDX
  - [ ] DIDX
  - [ ] Upload Data
    - [ ] Unencrypted, Uncompressed
    - [ ] Encrypted
    - [ ] Compressed
    - [ ] Blob Data
- [ ] Maybe in the future
  - [ ] Windows Support


## Format Overview
The PXAR format is a tar-like archive format used by Proxmox Backup Server (PBS) to store backups. It is a custom format that is not compatible with the standard tar format.

### Example Directory
For this example, this is the directory structure used to explain. The top-level directory is owned by root, and the files are owned by uid/gid 1000

- test-enc
  - abcdir
    - defdir
      - file4.txt
    - file2.txt
    - file3.txt
  - file.txt
  - symlink2abcdir (-> abcdir)

### Summary Structure
```
PXAR_ENTRY (root dir)
  PXAR_FILENAME (abcdir)
  PXAR_ENTRY (abcdir)
    PXAR_FILENAME (defdir)
    PXAR_ENTRY (defdir)
      PXAR_FILENAME (file4.txt)
      PXAR_ENTRY (file4.txt)
      PXAR_PAYLOAD (file4.txt contents)
      PXAR_GOODBYE (defdir, contains file4.txt)
    PXAR_FILENAME (file2.txt)
    PXAR_ENTRY (file2.txt)
    PXAR_PAYLOAD (file2.txt contents)
    PXAR_FILENAME (file3.txt)
    PXAR_ENTRY (file3.txt)
    PXAR_PAYLOAD (file3.txt contents)
    PXAR_GOODBYE (abcdir, contains defdir, file2.txt, file3.txt)
  PXAR_FILENAME (file.txt)
  PXAR_ENTRY (file.txt)
  PXAR_PAYLOAD (file.txt contents)
  PXAR_FILENAME (symlink2abcdir)
  PXAR_ENTRY (symlink2abcdir)
  PXAR_SYMLINK (symlink2abcdir -> abcdir)
  PXAR_GOODBYE (root dir, contains abcdir, file.txt, symlink2abcdir)
PXAR_GOODBYE_SPECIAL (special goodbye record for the file)

```

### Record
A record is composed of a header and a body. The header consists of the entry type and length of the body

```
Header:
  Type uint64
  Length uint64
```

For example, for the PXAR_ENTRY record, the header looks like this:

```
Type uint64: 0xd5956474e588acef
Length uint64: 56 (Header: 16 bytes, Body: 40 bytes)
```

### Start Record
The first record is the PXAR_ENTRY record for the top level directory. There is no filename associated with this record.
```
PXAR_ENTRY (0xd5956474e588acef, length 16 (header) + 40 (body))
  Mode uint64: 0o40755
  Flags uint64: 0x0
  Uid uint32: 0x0
  Gid uint32: 0x0
  MtimeSecs uint64: 1720277103
  MtimeNanos uint32: 123456789
  MtimePadding uint32: 0x0
```

### Directory contents
After the initial opening record are the records for the contents of the top-level directory. Generally, all record sets for a node consist of the following two sections followed by a payload section for the type of node.:

```
# The filename of the node
PXAR_FILENAME (0x16701121063917b3, length 16 (header) + filename length (6) + terminator byte (1))
  Filename string: "abcdir"
  Terminator uint8: 0x0

# The mode, owner and modified time information for the node
PXAR_ENTRY (0xd5956474e588acef, length 16 (header) + 40 (body))
  Mode uint64: 0o40755
  Flags uint64: 0x0
  Uid uint32: 0x0
  Gid uint32: 0x0
  MtimeSecs uint64: 1720277103
  MtimeNanos uint32: 123456789
  MtimePadding uint32: 0x0

# The payload section
# This consists of the PXAR_PAYLOAD for regular files, PXAR_SYMLINK for symlinks, etc.
# For directories, the payload is simply the records of the nodes in the dir terminated by the PXAR_GOODBYE record
```

### File Record
For the file4.txt-file in the example, the record looks like this:
```
# The filename
PXAR_FILENAME (0x16701121063917b3, length 16 (header) + filename length (9) + terminator byte (1))
  Filename string: "file4.txt"
  Terminator uint8: 0x0

# The metadata
PXAR_ENTRY (0xd5956474e588acef, length 16 (header) + 40 (body))
  Mode uint64: 0o100644
  Flags uint64: 0x0
  Uid uint32: 1000
  Gid uint32: 1000
  MtimeSecs uint64: 1720277103
  MtimeNanos uint32: 123456789
  MtimePadding uint32: 0x0

# The payload
PXAR_PAYLOAD (0x16701121063917b3, length 16 (header) + 22 (body))
  Payload string: "file4.txt content data"

```

### Symlink Record
For the symlink2abcdir symlink in the example, the record looks like this:

```
# The filename
PXAR_FILENAME (0x16701121063917b3, length 16 (header) + filename length (15) + terminator byte (1))
  Filename string: "symlink2abcdir"
  Terminator uint8: 0x0

# The metadata
PXAR_ENTRY (0xd5956474e588acef, length 16 (header) + 40 (body))
  Mode uint64: 0o120777
  Flags uint64: 0x0
  Uid uint32: 1000
  Gid uint32: 1000
  MtimeSecs uint64: 1720277103
  MtimeNanos uint32: 123456789
  MtimePadding uint32: 0x0

# The payload
PXAR_SYMLINK (0x16701121063917b3, length 16 (header) + 6 (body) + 1 (terminator))
  Target string: "abcdir"
  Terminator uint8: 0x0
```


### Goodbye Record
The goodbye record is a list of files that are in the directory sorted for more efficient searching. The process of creating the goodbye record is: 

1. Sort the list by hash
2. Create a binary search tree and adjust the offsets
3. Write the goodbye record

Each goodbye item looks like this:
```
GOODBYE_ITEM (no header):
  Hash uint64: The filename, hashed with siphash
  Offset uint64: How far back the start of the current node is in the archive (positive number)
  Length uint64: How long the written data for the node in question is
```

A goodbye record for the defdir directory would look like this:

```
PXAR_GOODBYE (0x2fec4fa642d5731d, length 16 (header) + (24 * (items)) + 24 (Goodbye tail marker))
  GOODBYE_ITEM (no header):
    Hash uint64: 0x2d3b7e7e1f2f1f7d
    Offset uint64: Difference between the start of the goodbye record and the start of the previous file4.txt record
    Length uint64: The length of the previous file4.txt record

  GOODBYE_ITEM (no header, tail marker for the folder)
    Hash uint64: 0xef5eed5b753e1555
    Offset uint64: Difference between the start of the goodbye Record and start of the folder it's referencing
    Length uint64: The length from the PXAR_GOODBYE header
```

### Full Example
This is a full example of the directory structure above. All the files contain the string "[filename] content data" where X is the number of the file.

```
PXAR_ENTRY (root dir) {
  Header: RecordType uint64: 0xd5956474e588acef
  Header: RecordLength uint64: 56
  Mode uint64: 0o40777
  Flags uint64: 0x0
  Uid uint32: 0x0
  Gid uint32: 0x0
  MtimeSecs uint64: 1720277103
  MtimeNanos uint32: 123456789
  MtimePadding uint32: 0x0
}
  PXAR_FILENAME (abcdir) {
    Header: RecordType uint64: 0x16701121063917b3
    Header: RecordLength uint64: 23
    Filename string: "abcdir"
    Terminator uint8: 0x0
  }

  PXAR_ENTRY (abcdir) (
    Header: RecordType uint64: 0xd5956474e588acef
    Header: RecordLength uint64: 56
    Mode uint64: 0o40755
    Flags uint64: 0x0
    Uid uint32: 1000
    Gid uint32: 1000
    MtimeSecs uint64: 1720277103
    MtimeNanos uint32: 123456789
    MtimePadding uint32: 0x0
  )

    PXAR_FILENAME (defdir) {
      Header: RecordType uint64: 0x16701121063917b3
      Header: RecordLength uint64: 23
      Filename string: "defdir"
      Terminator uint8: 0x0
    }

    PXAR_ENTRY (defdir) (
      Header: RecordType uint64: 0xd5956474e588acef
      Header: RecordLength uint64: 56
      Mode uint64: 0o40755
      Flags uint64: 0x0
      Uid uint32: 1000
      Gid uint32: 1000
      MtimeSecs uint64: 1720277103
      MtimeNanos uint32: 123456789
      MtimePadding uint32: 0x0
    )

      PXAR_FILENAME (file4.txt) {
        Header: RecordType uint64: 0x16701121063917b3
        Header: RecordLength uint64: 26 (16 + len(file4.txt) + 1)
        Filename string: "file4.txt"
        Terminator uint8: 0x0
      }

      PXAR_ENTRY (file4.txt) {
        Header: RecordType uint64: 0xd5956474e588acef
        Header: RecordLength uint64: 56
        Mode uint64: 0o100644
        Flags uint64: 0x0
        Uid uint32: 1000
        Gid uint32: 1000
        MtimeSecs uint64: 1720277103
        MtimeNanos uint32: 123456789
        MtimePadding uint32: 0x0
      }

      PXAR_PAYLOAD (file4.txt) {
        Header: RecordType uint64: 0x16701121063917b3
        Header: RecordLength uint64: 36 (16 + len(file4.txt content data))
        Payload string: "file-4-contents.txt\n"
      }

      PXAR_GOODBYE (defdir) {
        Header: RecordType uint64: 0x2fec4fa642d5731d
        Header: RecordLength uint64: 64 (16 + (24 * 1 item) + 24)

        GOODBYE_ITEM (file4.txt) {
          Hash uint64: 0xa303c432d543099 (siphash for file4.txt)
          Offset uint64: 120 (points to where PXAR_FILENAME (file4.txt starts in the byte stream, in this case 120 characters back))
          Length uint64: 120 The length of PXAR_FILENAME, PXAR_ENTRY, PXAR_PAYLOAD for file4.txt
        }
        
        GOODBYE_ITEM (defdir) {
          Hash uint64: 0xef5eed5b753e1555
          Offset uint64: 176 (points to where PXAR_GOODBYE (defdir) starts in the byte stream, in this case 176 characters back)
          Length uint64: 64 (16 + (24 * 1 item) + 24)
        }
      }
    
    PXAR_FILENAME (file2.txt) {
      Header: RecordType uint64: 0x16701121063917b3
      Header: RecordLength uint64: 26 (16 + len(file2.txt) + 1)
      Filename string: "file2.txt"
      Terminator uint8: 0x0
    }
    PXAR_ENTRY (file2.txt) {
      Header: RecordType uint64: 0xd5956474e588acef
      Header: RecordLength uint64: 56
      Mode uint64: 0o100644
      Flags uint64: 0x0
      Uid uint32: 1000
      Gid uint32: 1000
      MtimeSecs uint64: 1720277103
      MtimeNanos uint32: 123456789
      MtimePadding uint32: 0x0
    }
    PXAR_PAYLOAD (file2.txt) {
      Header: RecordType uint64: 0x16701121063917b3
      Header: RecordLength uint64: 36 (16 + len(file2.txt content data))
      Payload string: "file-2-contents.txt\n"
    }
    
    PXAR_FILENAME (file3.txt) {
      Header: RecordType uint64: 0x16701121063917b3
      Header: RecordLength uint64: 26 (16 + len(file3.txt) + 1)
      Filename string: "file3.txt"
      Terminator uint8: 0x0
    }
    PXAR_ENTRY (file3.txt) {
      Header: RecordType uint64: 0xd5956474e588acef
      Header: RecordLength uint64: 56
      Mode uint64: 0o100644
      Flags uint64: 0x0
      Uid uint32: 1000
      Gid uint32: 1000
      MtimeSecs uint64: 1720277103
      MtimeNanos uint32: 123456789
      MtimePadding uint32: 0x0
    }
    PXAR_PAYLOAD (file3.txt) {
      Header: RecordType uint64: 0x16701121063917b3
      Header: RecordLength uint64: 36 (16 + len(file3.txt content data))
      Payload string: "file-3-contents.txt\n"
    }

    PXAR_GOODBYE (defdir) {
      Header: RecordType uint64: 0x2fec4fa642d5731d
      Header: RecordLength uint64: 112 (16 + (24 * 3 items) + 24)

      GOODBYE_ITEM (defdir) {
        Hash uint64: 0x84cb159c0f774bc5
        Offset uint64: 503 (points to where PXAR_FILENAME (defdir)starts))
        Length uint64: 263 The length of PXAR_FILENAME, PXAR_ENTRY and children for defdir
      }

      GOODBYE_ITEM (file2.txt) {
        Hash uint64: 0x4a05703f633e1c8b
        Offset uint64: 240 (points to where PXAR_FILENAME (file2.txt starts in the byte stream, in this case 240 characters back))
        Length uint64: 120 The length of PXAR_FILENAME, PXAR_ENTRY, PXAR_PAYLOAD
      }

      GOODBYE_ITEM (file3.txt) {
        Hash uint64: 0x9b56a17c39e4818b
        Offset uint64: 120 (points to where PXAR_FILENAME (file3.txt starts in the byte stream, in this case 120 characters back))
        Length uint64: 120 The length of PXAR_FILENAME, PXAR_ENTRY, PXAR_PAYLOAD
      }
      
      GOODBYE_ITEM (abcdir) {
        Hash uint64: 0xef5eed5b753e1555
        Offset uint64: 559 (points to where PXAR_GOODBYE (abcdir) starts in the byte stream, in this case 176 characters back)
        Length uint64: 112 (16 + (24 * 3 items) + 24)
      }
    }

  PXAR_FILENAME (file.txt) {
    Header: RecordType uint64: 0x16701121063917b3
    Header: RecordLength uint64: 26 (16 + len(file.txt) + 1)
    Filename string: "file.txt"
    Terminator uint8: 0x0
  }

  PXAR_ENTRY (file.txt) {
    Header: RecordType uint64: 0xd5956474e588acef
    Header: RecordLength uint64: 56
    Mode uint64: 0o100644
    Flags uint64: 0x0
    Uid uint32: 1000
    Gid uint32: 1000
    MtimeSecs uint64: 1720277103
    MtimeNanos uint32: 123456789
    MtimePadding uint32: 0x0
  }

  PXAR_PAYLOAD (file.txt) {
    Header: RecordType uint64: 0x16701121063917b3
    Header: RecordLength uint64: 33 (16 + len(file.txt content data))
    Payload string: "file-txt-contents"
  }

  PXAR_FILENAME (symlink2abcdir) {
    Header: RecordType uint64: 0x16701121063917b3
    Header: RecordLength uint64: 32 (16 + len(symlink2abcdir) + 1)
    Filename string: "symlink2abcdir"
    Terminator uint8: 0x0
  }

  PXAR_ENTRY (symlink2abcdir) {
    Header: RecordType uint64: 0xd5956474e588acef
    Header: RecordLength uint64: 56
    Mode uint64: 0o120777
    Flags uint64: 0x0
    Uid uint32: 1000
    Gid uint32: 1000
    MtimeSecs uint64: 1720277103
    MtimeNanos uint32: 123456789
    MtimePadding uint32: 0x0
  }

  PXAR_SYMLINK (symlink2abcdir) {
    Header: RecordType uint64: 0x16701121063917b3
    Header: RecordLength uint64: 23 (16 + len(abcdir) + 1)
    Target string: "abcdir"
  }

  PXAR_GOODBYE (root directory) {
    Header: RecordType uint64: 0x2fec4fa642d5731d
    Header: RecordLength uint64: 112 (16 + (24 * 3 items) + 24)

    GOODBYE_ITEM (abcdir) {
      Hash uint64: 0x84cb159c0f774bc5
      Offset uint64: 919 (points to where PXAR_FILENAME (abcdir)starts))
      Length uint64: 694 The length of PXAR_FILENAME, PXAR_ENTRY and children for abcdir
    }

    GOODBYE_ITEM (file2.txt) {
      Hash uint64: 0x4a05703f633e1c8b
      Offset uint64: 110 (points to where PXAR_FILENAME (file.txt) starts
      Length uint64: 110 The length of PXAR_FILENAME, PXAR_ENTRY, PXAR_PAYLOAD
    }

    GOODBYE_ITEM (symlink2abcdir) {
      Hash uint64: 0x9b56a17c39e4818b
      Offset uint64: 975 (points to where PXAR_FILENAME (file3.txt starts in the byte stream, in this case 120 characters back))
      Length uint64: 115 The length of PXAR_FILENAME, PXAR_ENTRY, PXAR_SYMLINK
    }
    
    GOODBYE_ITEM (root directory) {
      Hash uint64: 0xef5eed5b753e1555
      Offset uint64: 559 (points to where PXAR_GOODBYE (rootdir) starts in the byte stream, in this case 112 characters back)
      Length uint64: 112 (16 + (24 * 3 items) + 24)
    }
  }

```