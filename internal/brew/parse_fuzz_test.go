package brew

import (
	"encoding/json"
	"testing"
)

func FuzzParseFormula(f *testing.F) {
	f.Add(`{"name":"test","versions":{"stable":"1.0"},"installed":[{"version":"1.0","installed_on_request":true,"installed_as_dependency":false,"time":1000000}]}`)
	f.Add(`{"name":""}`)
	f.Add(`{}`)
	f.Add(`{"name":"a","versions":{},"installed":null}`)
	f.Add(`{"name":"a","bottle":{"stable":{"files":null}}}`)

	f.Fuzz(func(t *testing.T, data string) {
		var fj formulaJSON
		if err := json.Unmarshal([]byte(data), &fj); err != nil {
			return
		}
		_ = parseFormula(fj)
	})
}

func FuzzParseCask(f *testing.F) {
	f.Add(`{"name":"test-cask","version":"1.0"}`)
	f.Add(`{}`)
	f.Add(`{"name":"a","artifacts":null}`)

	f.Fuzz(func(t *testing.T, data string) {
		var cj []byte
		cj = []byte(data)
		var result struct {
			Casks []struct {
				Name      string        `json:"name"`
				Version   string        `json:"version"`
				Installed []interface{} `json:"installed"`
				Artifacts []interface{} `json:"artifacts"`
			} `json:"casks"`
		}
		if err := json.Unmarshal(cj, &result); err != nil {
			return
		}
	})
}
