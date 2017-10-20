package riddick

import "testing"
import "os"

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

	o, s, err := a.header()
	if err != nil {
		t.Fatal(err)
	}
	if s != 2048 {
		t.Errorf("expected %d got %d", 2048, s)
	}
	if o != 2048 {
		t.Errorf("expected %d got %d", 2048, o)
	}
}
