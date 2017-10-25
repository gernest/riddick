package riddick

import "testing"
import "os"
import "fmt"

// fixture/SampleDSStore
// two files
// => foo.txt
// => bar.txt
// One empty directory
// => nothing
func TestAllocator_header(t *testing.T) {
	f, err := os.Open("fixture/SampleDSStore")
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
		t.Errorf("expected %d got %d", 2048, o)
	}
}

func TestAllocator(t *testing.T) {
	f, err := os.Open("fixture/SampleDSStore")
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

	b, err := a.GetBlock(2)
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.entry()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(*e)
}
