package riddick

import (
	"encoding/binary"
	"errors"
	"os"
)

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

func (a *allocator) GetBlock(bid uint32) (*block, error) {
	if len(a.offsets) <= int(bid) {
		return nil, errors.New("Cannot find key in Offset-Table")
	}
	addr := a.offsets[bid]
	offset := int(addr) & ^0x1f
	size := 1 << (uint(addr) & 0x1f)
	block, err := newBlock(a, uint32(offset), uint32(size)) ///+4??
	if err != nil {
		return nil, errors.New("Cannot create/read block")
	}
	return block, nil
}

type entry struct {
	filename string
	code     string
	typeCode string
	value    interface{}
}

func (a *allocator) traverse(block uint32, f func(*entry) error) error {
	node, err := a.GetBlock(block)
	if err != nil {
		return err
	}
	nextPtr, err := node.readUint32()
	if err != nil {
		return err
	}
	count, err := node.readUint32()
	if err != nil {
		return err
	}
	if nextPtr > 0 {
		//This may be broken
		for i := 0; i < int(count); i++ {
			next, err := node.readUint32()
			if err != nil {
				return err
			}
			err = a.traverse(next, f)
			if err != nil {
				return err
			}

			e, err := node.entry()
			if err != nil {
				return err
			}
			if err = f(e); err != nil {
				return err
			}
		}
		err := a.traverse(nextPtr, f)
		if err != nil {
			return err
		}
	} else {
		for i := 0; i < int(count); i++ {
			e, err := node.entry()
			if err = f(e); err != nil {
				return err
			}
		}
	}
	return nil
}
