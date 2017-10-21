package riddick

import "testing"
import "os"

func TestAllocator(t *testing.T) {
	f, err := os.Open("fixture/Test_DS_Store")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = newAllocator(f)
	if err != nil {
		t.Fatal(err)
	}
}
