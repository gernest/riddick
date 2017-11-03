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
		a:      a,
		offset: offset,
		size:   size,
		data:   v,
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

func (b *block) entry() (*entry, error) {
	e := &entry{}
	if err := b.readToEntry(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (b *block) readToEntry(e *entry) error {
	n, err := b.filaname()
	if err != nil {
		return err
	}
	e.filename = n
	code, err := b.code()
	if err != nil {
		return err
	}
	e.code = code
	typeCode, err := b.typeCode()
	if err != nil {
		return err
	}
	e.typeCode = typeCode

	bytesToSkip := -1

	switch typeCode {
	case "bool":
		bytesToSkip = 1
	case "type", "long", "shor":
		bytesToSkip = 4
	case "comp", "dutc":
		bytesToSkip = 8
	case "blob":
		blen, err := b.readUint32()
		if err != nil {
			return err
		}
		bytesToSkip = int(blen)
	case "ustr":
		blen, err := b.readUint32()
		if err != nil {
			return err
		}
		bytesToSkip = int(2 * blen)
	}
	if bytesToSkip <= 0 {
		return errors.New("Unknown file format")
	}
	o, err := b.readBuf(bytesToSkip)
	if err != nil {
		return err
	}
	e.data = o
	return nil
}

func (b *block) filaname() (string, error) {
	length, err := b.readUint32()
	if err != nil {
		return "", err
	}
	buf, err := b.readBuf(int(2 * length))
	if err != nil {
		return "", err
	}
	n, err := utf16beDec.Bytes(buf)
	if err != nil {
		return "", err
	}
	return string(n), nil
}

func (b *block) string(size int) (string, error) {
	o, err := b.readBuf(size)
	if err != nil {
		return "", err
	}
	return string(o), nil
}

func (b *block) code() (string, error) {
	return b.string(4)
}

func (b *block) typeCode() (string, error) {
	return b.string(4)
}

func (b *block) seek(pos int, whence int) error {
	switch whence {
	case os.SEEK_CUR:
		pos += b.pos
	case os.SEEK_END:
		pos = int(b.size) - pos
	}
	if pos < 0 || pos > int(b.size) {
		return errors.New("seek out of range")
	}
	b.pos = pos
	return nil
}

func (b *block) Write(data []byte) (int, error) {
	if b.pos+len(data) > int(b.size) {
		return 0, errors.New("trying to write past end of block")
	}
	copy(b.data[b.pos:b.pos+len(data)], data)
	b.pos += len(data)
	b.dirty = true
	return len(data), nil
}

func (b *block) flush() error {
	if b.dirty {
		_, err := b.a.write(int(b.offset), b.data)
		if err != nil {
			return err
		}
	}
	return nil
}
