package riddick

import (
	"os"
	"reflect"
	"testing"
)

// fixture/SampleDSStore
// two files
// => foo.txt
// => bar.txt
// One empty directory
// => nothing
func TestAllocator_header(t *testing.T) {
	f, err := os.Open("fixture/Test_DS_Store")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a := &allocator{
		file: f,
	}

	o, s, err := a.header()
	if err != nil {
		t.Fatal(err)
	}
	if s != 2048 {
		t.Errorf("expected %d got %d", 2048, s)
	}
	if o != 4096 {
		t.Errorf("expected %d got %d", 4096, o)
	}
}

func TestAllocator(t *testing.T) {
	f, err := os.Open("fixture/Test_DS_Store")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := newAllocator(f)
	if err != nil {
		t.Fatal(err)
	}
	k := a.toc["DSDB"]
	if k != 1 {
		t.Errorf("expected 1 got %d", k)
	}
	o := []uint32{4107, 69, 135}
	if !reflect.DeepEqual(o, a.offsets) {
		t.Errorf("expected %v got %v", o, a.offsets)
	}
}
func TestAllocator_Entries(t *testing.T) {
	f, err := os.Open("fixture/SampleDSStore")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := newAllocator(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := a.GetBlock(1)
	if err != nil {
		t.Fatal(err)
	}
	if b.offset != 64 {
		t.Errorf("expected %d got %d", 64, b.offset)
	}
	if b.size != 32 {
		t.Errorf("expected %d tot %d", 32, b.size)
	}
	err = a.traverse(1, func(e *entry) error {
		switch e.filename {
		case "bar.txt":
			switch e.code {
			case "Iloc":
				x, y, err := e.decodeIloc()
				if err != nil {
					return err
				}
				if x != 59 {
					t.Errorf("expected 59 got %d", x)
				}
				if y != 40 {
					t.Errorf("expected 40 got %d", y)
				}
			}
		case "foo.txt":
			switch e.code {
			case "Iloc":
				x, y, err := e.decodeIloc()
				if err != nil {
					return err
				}
				if x != 169 {
					t.Errorf("expected 169 got %d", x)
				}
				if y != 40 {
					t.Errorf("expected 40 got %d", y)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
