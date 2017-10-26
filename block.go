package riddick

import (
	"encoding/binary"
	"io"
	"reflect"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
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
	n, err := b.filaname()
	if err != nil {
		return nil, err
	}
	e.filename = n
	code, err := b.code()
	if err != nil {
		return nil, err
	}
	e.code = code
	typeCode, err := b.typeCode()
	if err != nil {
		return nil, err
	}
	e.typeCode = typeCode
	switch typeCode {

	}
	return e, nil
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
	n := utf16be2utf8(buf)
	return n, nil
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

func utf16be2utf8(utf16be []byte) string {
	//Taken from http://play.golang.org/p/xtG1e9iqA1
	n := len(utf16be)
	// Convert to []uint16
	// hop through unsafe to skip any actual allocation/copying
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&utf16be))
	header.Len /= 2
	shorts := *(*[]uint16)(unsafe.Pointer(&header))
	// shorts may need byte-swapping
	for i := 0; i < n; i += 2 {
		shorts[i/2] = (uint16(utf16be[i]) << 8) | uint16(utf16be[i+1])
	}

	// Convert to []byte
	count := 0
	for i := 0; i < len(shorts); i++ {
		r := rune(shorts[i])
		if utf16.IsSurrogate(r) {
			i++
			r = utf16.DecodeRune(r, rune(shorts[i]))
		}
		count += utf8.RuneLen(r)
	}
	buf := make([]byte, count)
	bi := 0
	for i := 0; i < len(shorts); i++ {
		r := rune(shorts[i])
		if utf16.IsSurrogate(r) {
			i++
			r = utf16.DecodeRune(r, rune(shorts[i]))
		}
		bi += utf8.EncodeRune(buf[bi:], r)
	}
	return string(buf)
}
