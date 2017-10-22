package riddick

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type block struct {
	a            *allocator
	offset, size uint32
	pos          int
	data         []byte
	dirty        bool
}

func newBlock(a *allocator, offset uint32, size uint32) (*block, error) {
	v, err := a.read(int64(offset), int(size))
	if err != nil {
		return nil, err
	}
	return &block{
		a:    a,
		size: size,
		data: v,
	}, nil
}

func (b *block) readUint32() (uint32, error) {
	var v uint32
	err := binary.Read(b, binary.BigEndian, &v)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (b *block) ReadByte() (byte, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	o := b.data[b.pos]
	b.pos++
	return o, nil
}

func (b *block) Read(v []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(v, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *block) uint32Slice(size int) ([]uint32, error) {
	a := make([]uint32, size)
	err := binary.Read(b, binary.BigEndian, a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (b *block) readByte() (byte, error) {
	var v byte
	err := binary.Read(b, binary.BigEndian, &v)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (b *block) readBuf(length int) (buf []byte, err error) {
	buf = make([]byte, length)
	_, err = b.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (b *block) skip(i int) {
	b.pos += i
}

type allocator struct {
	file     *os.File
	root     *block
	pos      int
	offsets  []uint32
	toc      map[string]uint32
	freeList map[uint32][]uint32
	unkown1  string
	unkown2  uint32
}

func newAllocator(f *os.File) (*allocator, error) {
	a := &allocator{
		file:     f,
		toc:      make(map[string]uint32),
		freeList: make(map[uint32][]uint32),
	}
	offset, size, err := a.header()
	if err != nil {
		return nil, err
	}
	r, err := newBlock(a, offset, size)
	if err != nil {
		return nil, err
	}
	a.root = r
	return a.init()
}

func (a *allocator) init() (*allocator, error) {
	if err := a.readOffsets(); err != nil {
		return nil, err
	}
	if err := a.readToc(); err != nil {
		return nil, err
	}
	if err := a.readFreeList(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *allocator) header() (uint32, uint32, error) {
	m, err := a.uint32()
	if err != nil {
		return 0, 0, err
	}
	if m != 1 {
		return 0, 0, errors.New("Not a buddy file")
	}
	magic, err := a.string(4)
	if err != nil {
		return 0, 0, err
	}
	if string(magic) != "Bud1" {
		return 0, 0, errors.New("Not a buddy file")
	}
	o, err := a.uint32()
	if err != nil {
		return 0, 0, err
	}
	s, err := a.uint32()
	if err != nil {
		return 0, 0, err
	}
	o2, err := a.uint32()
	if err != nil {
		return 0, 0, err
	}
	u1, err := a.string(16) // Unknown1
	if err != nil {
		return 0, 0, err
	}
	a.unkown1 = u1
	if o != o2 {
		return 0, 0, errors.New("Root addresses differ")
	}
	return o, s, nil
}

func (a *allocator) skip() error {
	o := a.pos + 4
	_, err := a.file.Seek(int64(o), os.SEEK_SET)
	if err != nil {
		return err
	}
	return nil
}

func (a *allocator) uint32() (uint32, error) {
	var ab [4]byte
	b := ab[:]
	size, err := a.file.Read(b)
	if err != nil {
		return 0, err
	}
	a.pos += size
	return binary.BigEndian.Uint32(b), nil
}

func (a *allocator) string(size int) (string, error) {
	v := make([]byte, size)
	n, err := a.file.Read(v)
	if err != nil {
		return "", err
	}
	a.pos += n
	return string(v), nil
}

func (a *allocator) read(offset int64, size int) ([]byte, error) {
	o := make([]byte, size)
	_, err := a.file.ReadAt(o, offset+4)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (a *allocator) readOffsets() error {
	count, err := a.root.readUint32()
	if err != nil {
		return err
	}
	u2, err := a.root.readUint32()
	if err != nil {
		return err
	}
	a.unkown2 = u2
	c := (int(count) + 255) & ^255
	for c != 0 {
		o, err := a.root.uint32Slice(256)
		if err != nil {
			return err
		}
		a.offsets = append(a.offsets, o...)
		c -= 256
	}
	a.offsets = a.offsets[:count]
	return nil
}

func (a *allocator) readToc() error {
	toccount, err := a.root.readUint32()
	if err != nil {
		return err
	}
	for i := toccount; i > 0; i-- {
		tlen, err := a.root.readByte()
		if err != nil {
			return err
		}
		name, err := a.root.readBuf(int(tlen))
		if err != nil {
			return err
		}
		value, err := a.root.readUint32()
		if err != nil {
			return err
		}
		a.toc[string(name)] = value
	}
	return nil
}

func (a *allocator) readFreeList() error {
	for i := 0; i < 32; i++ {
		blkcount, err := a.root.readUint32()
		if err != nil {
			return err
		}
		if blkcount == 0 {
			continue
		}
		a.freeList[uint32(i)] = make([]uint32, 0)
		for k := 0; k < int(blkcount); k++ {
			val, err := a.root.readUint32()
			if err != nil {
				return err
			}
			if val == 0 {
				continue
			}
			a.freeList[uint32(i)] = append(a.freeList[uint32(i)], val)
		}
	}
	return nil
}
