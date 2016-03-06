package structure

import "testing"

func TestParse(t *testing.T) {
	testFilePath := "/Users/achiku/tmp/schema.json"
	st, err := JSONParse(testFilePath)
	if err != nil {
		t.Error(err)
	}
	for _, s := range st {
		t.Logf("\n%s", s.String(true))
	}
}
