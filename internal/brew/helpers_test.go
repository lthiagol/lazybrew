package brew

import (
	"reflect"
	"testing"
)

func TestDecodeEachKeepsEntriesThatPass(t *testing.T) {
	data := []byte(`[{"name":"a"},"bad",{"name":""},{"name":"b"}]`)
	got, err := decodeEach(data, func(s Service) bool { return s.Name != "" })
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, s := range got {
		names = append(names, s.Name)
	}
	if !reflect.DeepEqual(names, []string{"a", "b"}) {
		t.Errorf("names = %q, want [a b]", names)
	}
}

func TestDecodeEachRejectsOutputThatIsNotAnArray(t *testing.T) {
	if _, err := decodeEach([]byte(`{}`), func(s Service) bool { return true }); err == nil {
		t.Error("expected an error for an object")
	}
}

func TestNonEmptyLinesTrimsAndDropsBlanks(t *testing.T) {
	got := nonEmptyLines([]byte("\n  a  \n\n\tb\r\n \n"))
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("nonEmptyLines = %q, want [a b]", got)
	}
}
