package riddick

import (
	"encoding/binary"
	"errors"
	"os"
)

type block struct {
	a                 *allocator
	offset, size, pos uint32
	data              []byte
	dirty             bool
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

type allocator struct {
	file    *os.File
	root    *block
	pos     int
	offsets []uint32
	toc     map[string]uint32
	free    map[uint32][]uint32
}

func newAllocator(f *os.File) (*allocator, error) {
	a := &allocator{
		file: f,
		toc:  make(map[string]uint32),
		free: make(map[uint32][]uint32),
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
	if o != o2 {
		return 0, 0, errors.New("Root addresses differ")
	}
	return o, s, nil
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
