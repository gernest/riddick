package riddick

import (
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	f, err := os.Open("fixture/SampleDSStore")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := newAllocator(f)
	if err != nil {
		t.Fatal(err)
	}

	s, err := newStore(a)
	if err != nil {
		t.Fatal(err)
	}

	v, err := s.find("foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	i := v[0]
	if i.filename != "foo.txt" {
		t.Errorf("expected foo.txt got %s", i.filename)
	}
}
