package data

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"runtime"
	"sync"

	"github.com/sylabs/squashfs/internal/decompress"
	"github.com/sylabs/squashfs/internal/toreader"
)

type FragReaderConstructor func() (io.Reader, error)

type FullReader struct {
	r              io.ReaderAt
	d              decompress.Decompressor
	frag           FragReaderConstructor
	retPool        *sync.Pool
	sizes          []uint32
	initialOffset  int64
	finalBlockSize uint64
	blockSize      uint32
	goroutineLimit uint16
}

func NewFullReader(r io.ReaderAt, initialOffset int64, d decompress.Decompressor, sizes []uint32, finalBlockSize uint64, blockSize uint32) *FullReader {
	return &FullReader{
		r:              r,
		d:              d,
		sizes:          sizes,
		initialOffset:  initialOffset,
		goroutineLimit: uint16(runtime.NumCPU()),
		finalBlockSize: finalBlockSize,
		blockSize:      blockSize,
		retPool: &sync.Pool{
			New: func() any {
				return &retValue{}
			},
		},
	}
}

func (r *FullReader) AddFrag(frag FragReaderConstructor) {
	r.frag = frag
}

func (r *FullReader) SetGoroutineLimit(limit uint16) {
	r.goroutineLimit = limit
}

type retValue struct {
	err   error
	data  []byte
	index uint64
}

func (r *FullReader) process(index uint64, fileOffset uint64, retChan chan *retValue) {
	ret := r.retPool.Get().(*retValue)
	ret.index = index
	realSize := r.sizes[index] &^ (1 << 24)
	if realSize == 0 {
		if index == uint64(len(r.sizes))-1 && r.frag == nil {
			ret.data = make([]byte, r.finalBlockSize)
		} else {
			ret.data = make([]byte, r.blockSize)
		}
		ret.err = nil
		retChan <- ret
		return
	}
	ret.data = make([]byte, realSize)
	ret.err = binary.Read(toreader.NewReader(r.r, int64(r.initialOffset)+int64(fileOffset)), binary.LittleEndian, &ret.data)
	if r.sizes[index] == realSize {
		ret.data, ret.err = r.d.Decompress(ret.data)
	}
	retChan <- ret
}

func (r *FullReader) WriteTo(w io.Writer) (int64, error) {
	var curIndex uint64
	var curOffset uint64
	var toProcess uint16
	var wrote int64
	cache := make(map[uint64]*retValue)
	var errCache []error
	retChan := make(chan *retValue, r.goroutineLimit)
	for i := uint64(0); i < uint64(math.Ceil(float64(len(r.sizes))/float64(r.goroutineLimit))); i++ {
		toProcess = uint16(len(r.sizes)) - (uint16(i) * r.goroutineLimit)
		if toProcess > r.goroutineLimit {
			toProcess = r.goroutineLimit
		}
		// Start all the goroutines
		for j := uint16(0); j < toProcess; j++ {
			go r.process((i*uint64(r.goroutineLimit))+uint64(j), curOffset, retChan)
			curOffset += uint64(r.sizes[(i*uint64(r.goroutineLimit))+uint64(j)]) &^ (1 << 24)
		}
		// Then consume the results on retChan
		for j := uint16(0); j < toProcess; j++ {
			res := <-retChan
			// If there's an error, we don't care about the results.
			if res.err != nil {
				errCache = append(errCache, res.err)
				if len(cache) > 0 {
					clear(cache)
				}
				continue
			}
			// If there has been an error previously, we don't care about the results.
			// We still want to wait for all the goroutines to prevent resources being wasted.
			if len(errCache) > 0 {
				continue
			}
			// If we don't need the data yet, we cache it and move on
			if res.index != curIndex {
				cache[res.index] = res
				continue
			}
			// If we do need the data, we write it
			wr, err := w.Write(res.data)
			wrote += int64(wr)
			if err != nil {
				errCache = append(errCache, err)
				if len(cache) > 0 {
					clear(cache)
				}
				continue
			}
			r.retPool.Put(res)
			curIndex++
			// Now we recursively try to clear the cache
			for len(cache) > 0 {
				res, ok := cache[curIndex]
				if !ok {
					break
				}
				wr, err := w.Write(res.data)
				wrote += int64(wr)
				if err != nil {
					errCache = append(errCache, err)
					if len(cache) > 0 {
						clear(cache)
					}
					break
				}
				delete(cache, curIndex)
				r.retPool.Put(res)
				curIndex++
			}
		}
		if len(errCache) > 0 {
			return wrote, errors.Join(errCache...)
		}
	}
	if r.frag != nil {
		rdr, err := r.frag()
		if err != nil {
			return wrote, err
		}
		wr, err := io.Copy(w, rdr)
		wrote += wr
		if l, ok := rdr.(*io.LimitedReader); ok {
			if cl, ok := l.R.(io.Closer); ok {
				cl.Close()
			}
		}
		if err != nil {
			return wrote, err
		}
	}
	return wrote, nil
}
