package regen

import "testing"

func TestGenerate(t *testing.T) {
	err := Generate("../testdata/cats.zip", 5, 64, false)
	if err != nil {
		t.Error(err)
	}
}
