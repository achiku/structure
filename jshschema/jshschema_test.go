package jshschema

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	testFilePath := "/Users/achiku/tmp/schema.json"
	f, err := os.Open(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	b, err := Generate(f, "testpackage")
	if err != nil {
		t.Error(err)
	}

	t.Log(b)
}
