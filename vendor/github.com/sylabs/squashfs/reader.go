package squashfs

import (
	"io"
	"time"

	squashfslow "github.com/sylabs/squashfs/low"
)

type Reader struct {
	*FS
	Low *squashfslow.Reader
}

func NewReader(r io.ReaderAt) (*Reader, error) {
	rdr, err := squashfslow.NewReader(r)
	if err != nil {
		return nil, err
	}
	out := &Reader{
		Low: rdr,
	}
	out.FS = &FS{
		d: rdr.Root,
		r: out,
	}
	return out, nil
}

func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.Low.Superblock.ModTime), 0)
}
