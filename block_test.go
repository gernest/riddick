package riddick

import "testing"

import "bytes"

func TestBlockWrite(t *testing.T) {
	s := "hello,world"
	r := " gernest"
	size := len(s) * 2
	b := &block{
		size: uint32(size),
		data: make([]byte, size),
	}

	n, err := b.Write([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(s) {
		t.Errorf("expected %d got %d", len(s), n)
	}
	if !bytes.HasPrefix(b.data, []byte(s)) {
		t.Errorf("expected %s to be written from the beginning of block", s)
	}

	n, err = b.Write([]byte(r))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(r) {
		t.Errorf("expected %d got %d", len(r), n)
	}
	if !bytes.HasPrefix(b.data, []byte(s+r)) {
		t.Errorf("expected %s to be written from the beginning of block", s+r)
	}
}
