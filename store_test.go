package riddick

import (
	"fmt"
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

	fmt.Println(s)
}
