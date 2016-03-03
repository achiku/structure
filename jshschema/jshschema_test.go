package jshschema

import "testing"

func TestNewGenerate(t *testing.T) {
	testFilePath := "/Users/achiku/tmp/schema.json"
	b, err := Generate(testFilePath, "testpackage")
	if err != nil {
		t.Error(err)
	}
	t.Log(b)
}
