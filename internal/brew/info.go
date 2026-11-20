package brew

import "encoding/json"

// ParseInfo decodes a `brew info --json=v2` payload into its formulae and
// casks, through the same parsers the readers use. A payload with only one of
// the two keys yields an empty list for the other.
func ParseInfo(data []byte) ([]Formula, []Cask, error) {
	var fj formulaeJSON
	if err := json.Unmarshal(data, &fj); err != nil {
		return nil, nil, err
	}
	var cj casksJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return nil, nil, err
	}

	formulae := make([]Formula, 0, len(fj.Formulae))
	for _, f := range fj.Formulae {
		formulae = append(formulae, parseFormula(f))
	}
	casks := make([]Cask, 0, len(cj.Casks))
	for _, c := range cj.Casks {
		casks = append(casks, parseCask(c))
	}
	return formulae, casks, nil
}
