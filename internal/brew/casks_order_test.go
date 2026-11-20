package brew

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestParseCaskArtifactsKeepSortedOrder: a cask's artifacts come out in the
// same order on every parse. They used to follow Go's random map order.
func TestParseCaskArtifactsKeepSortedOrder(t *testing.T) {
	raw := `{"name":"x","artifacts":[{"pkg":"x.pkg","binary":["bin/x"],"app":["X.app"],"colorpicker":["a"]}]}`
	want := []string{"X.app", "bin/x", "a", "x.pkg"}

	for i := 0; i < 20; i++ {
		var c caskJSON
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatal(err)
		}
		if got := parseCask(c).Artifacts; !reflect.DeepEqual(got, want) {
			t.Fatalf("parse %d: artifacts = %q, want %q", i, got, want)
		}
	}
}
