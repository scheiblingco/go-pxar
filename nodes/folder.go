package nodes

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/dchest/siphash"
	"github.com/scheiblingco/go-pxar/pxar"
)

func (ref *FolderRef) GetChildren() []NodeRef {
	return ref.Children
}

func (ref *FolderRef) GetHash() uint64 {
	return siphash.Hash(pxar.PXAR_HASH_KEY_1, pxar.PXAR_HASH_KEY_2, []byte(ref.Name))
}

func (ref *FolderRef) WriteCatalogue(buf *bytes.Buffer, pos *uint64, parentStartPos uint64) ([]byte, uint64, error) {
	startPos := *pos
	totalLen := uint64(0)
	// Add folder length
	table := bytes.NewBuffer([]byte{})

	lenUvi := MakeUvarint(uint64(len(ref.Children)))

	n, err := table.Write(lenUvi)
	if err != nil {
		return nil, 0, err
	}

	totalLen += uint64(n)
	*pos += uint64(n)

	for _, child := range ref.Children {
		chBytes, n, err := child.WriteCatalogue(buf, pos, startPos)
		if err != nil {
			return nil, 0, err
		}

		table.Write(chBytes)
		totalLen += uint64(len(chBytes))
		*pos += n + uint64(len(chBytes))
	}

	buf.Write(MakeUvarint(uint64(table.Len())))
	buf.Write(table.Bytes())

	selfItem := bytes.NewBuffer([]byte{})
	selfItem.WriteByte(byte(pxar.DirectoryEntry))
	selfItem.Write(MakeUvarint(uint64(len(ref.Name))))
	selfItem.WriteString(ref.Name)
	selfItem.Write(MakeUvarint(uint64(totalLen + 1)))

	return selfItem.Bytes(), totalLen, nil
}

func (ref *FolderRef) WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	startPos := *pos

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.Write(buf, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	for _, child := range ref.Children {
		n, err := child.WritePayload(buf, pos)
		if err != nil {
			return 0, err
		}

		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   child.GetHash(),
			Offset: *pos,
			Length: n,
		})
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}

func (ref *FolderRef) WritePayloadChannel(ch chan []byte, pos *uint64) (uint64, error) {
	startPos := *pos

	if len(ref.GoodbyeItems) > 0 {
		// Reset goodbye items in case file is written more than once
		ref.GoodbyeItems = []pxar.GoodbyeItem{}
	}

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.WriteChannel(ch, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	for _, child := range ref.Children {
		n, err := child.WritePayloadChannel(ch, pos)
		if err != nil {
			return 0, err
		}

		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   child.GetHash(),
			Offset: *pos,
			Length: n,
		})
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}
// WritePayloadAsync implements concurrent directory processing
func (ref *FolderRef) WritePayloadAsync(buf *bytes.Buffer, pos *uint64, workers int) (uint64, error) {
	startPos := *pos

	if len(ref.GoodbyeItems) > 0 {
		// Reset goodbye items in case folder is written more than once
		ref.GoodbyeItems = []pxar.GoodbyeItem{}
	}

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.Write(buf, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	// Process children with concurrency
	err = ref.processChildrenAsync(buf, pos, workers)
	if err != nil {
		return 0, err
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}

// WritePayloadChannelAsync implements concurrent directory processing for channels
func (ref *FolderRef) WritePayloadChannelAsync(ch chan []byte, pos *uint64, workers int) (uint64, error) {
	startPos := *pos

	if len(ref.GoodbyeItems) > 0 {
		// Reset goodbye items in case folder is written more than once
		ref.GoodbyeItems = []pxar.GoodbyeItem{}
	}

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.WriteChannel(ch, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	// Process children with concurrency
	err = ref.processChildrenChannelAsync(ch, pos, workers)
	if err != nil {
		return 0, err
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}

// Result structure for concurrent child processing
type childResult struct {
	index        int
	goodbyeItem  pxar.GoodbyeItem
	buffer       *bytes.Buffer
	err          error
}

// processChildrenAsync processes folder children concurrently for buffer output
func (ref *FolderRef) processChildrenAsync(buf *bytes.Buffer, pos *uint64, workers int) error {
	if len(ref.Children) == 0 {
		return nil
	}

	// Limit workers to number of children to avoid unnecessary goroutines
	if workers > len(ref.Children) {
		workers = len(ref.Children)
	}

	// Create channels for work distribution and results
	workChan := make(chan int, len(ref.Children))
	resultChan := make(chan childResult, len(ref.Children))
	
	// Fill work channel with child indices
	for i := range ref.Children {
		workChan <- i
	}
	close(workChan)

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for childIndex := range workChan {
				child := ref.Children[childIndex]
				childBuf := bytes.NewBuffer([]byte{})
				childPos := uint64(0)
				
				// Process child using async method if available, otherwise sync
				var n uint64
				var err error
				if folder, ok := child.(*FolderRef); ok {
					n, err = folder.WritePayloadAsync(childBuf, &childPos, workers)
				} else {
					n, err = child.WritePayload(childBuf, &childPos)
				}
				
				result := childResult{
					index:  childIndex,
					buffer: childBuf,
					err:    err,
				}
				
				if err == nil {
					// Store the length, offset will be calculated when writing
					result.goodbyeItem = pxar.GoodbyeItem{
						Hash:   child.GetHash(),
						Offset: 0, // Will be set when writing sequentially
						Length: n,
					}
				}
				
				resultChan <- result
			}
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and maintain order
	results := make([]childResult, len(ref.Children))
	for result := range resultChan {
		if result.err != nil {
			return result.err
		}
		results[result.index] = result
	}

	// Write children in original order and update positions
	for i, result := range results {
		// Write child buffer to main buffer
		n, err := buf.Write(result.buffer.Bytes())
		if err != nil {
			return err
		}
		
		*pos += uint64(n)
		
		// Create goodbye item with correct offset (points to end position)
		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   ref.Children[i].GetHash(),
			Offset: *pos,
			Length: result.goodbyeItem.Length,
		})
	}

	return nil
}

// processChildrenChannelAsync processes folder children concurrently for channel output
func (ref *FolderRef) processChildrenChannelAsync(ch chan []byte, pos *uint64, workers int) error {
	if len(ref.Children) == 0 {
		return nil
	}

	// Limit workers to number of children to avoid unnecessary goroutines
	if workers > len(ref.Children) {
		workers = len(ref.Children)
	}

	// Create channels for work distribution and results
	workChan := make(chan int, len(ref.Children))
	resultChan := make(chan childResult, len(ref.Children))
	
	// Fill work channel with child indices
	for i := range ref.Children {
		workChan <- i
	}
	close(workChan)

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for childIndex := range workChan {
				child := ref.Children[childIndex]
				childCh := make(chan []byte, 10)
				childPos := uint64(0)
				
				// Process child using async method if available, otherwise sync
				go func() {
					defer close(childCh)
					if folder, ok := child.(*FolderRef); ok {
						folder.WritePayloadChannelAsync(childCh, &childPos, workers)
					} else {
						child.WritePayloadChannel(childCh, &childPos)
					}
				}()
				
				// Collect all data from child channel
				var childBuf bytes.Buffer
				for data := range childCh {
					childBuf.Write(data)
				}
				
				result := childResult{
					index:  childIndex,
					buffer: &childBuf,
					goodbyeItem: pxar.GoodbyeItem{
						Hash:   child.GetHash(),
						Offset: 0, // Will be set when writing
						Length: uint64(childBuf.Len()),
					},
				}
				
				resultChan <- result
			}
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and maintain order
	results := make([]childResult, len(ref.Children))
	for result := range resultChan {
		if result.err != nil {
			return result.err
		}
		results[result.index] = result
	}

	// Write children in original order and update positions
	for i, result := range results {
		// Send child data to channel
		ch <- result.buffer.Bytes()
		
		*pos += uint64(result.buffer.Len())
		
		// Create goodbye item with correct offset
		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   ref.Children[i].GetHash(),
			Offset: *pos,
			Length: result.goodbyeItem.Length,
		})
	}

	return nil
}
