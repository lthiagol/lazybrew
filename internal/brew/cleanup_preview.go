package brew

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// CleanupPreview is what `brew cleanup -n` would remove, parsed from its output.
type CleanupPreview struct {
	// Items are the paths cleanup would remove or prune, in brew's order.
	Items []CleanupItem
	// Total is the size brew gives for the whole run, worded as brew words it
	// ("950MB"). It is "" when brew printed none.
	Total string
	// Unused are the formulae the run's autoremove step would uninstall: brew
	// cleanup runs autoremove too, unless HOMEBREW_NO_AUTOREMOVE is set.
	Unused []string
	// Raw are the lines no pattern matched, kept so none is dropped silently.
	Raw []string
}

// CleanupItem is one path that cleanup would remove or prune.
type CleanupItem struct {
	Path string
	// Kind is brew's label for a path it removes without a size: "broken link"
	// or "empty directory". It is "" for a file or directory with a size.
	Kind string
	// Files is the file count brew gives. A single file has no count, so it is
	// 1; a path with no size, or a prune with no size, gives 0 or its count.
	Files int
	// Size is brew's size text ("230.3MB"); "" when brew gives none.
	Size string
	// Bytes is Size in bytes, for ordering; 0 when Size does not parse.
	Bytes int64
}

// AutoremovePreview is what `brew autoremove -n` would uninstall.
type AutoremovePreview struct {
	// Formulae are the names brew would uninstall, sorted as brew sorts them.
	Formulae []string
	// Raw are the lines no pattern matched.
	Raw []string
}

var (
	cleanupTotalRE   = regexp.MustCompile(`^This operation would free approximately (\S+) of disk space\.$`)
	cleanupRemoveRE  = regexp.MustCompile(`^Would remove: (.+) \(([^()]+)\)$`)
	cleanupKindRE    = regexp.MustCompile(`^Would remove \((broken link|empty directory)\): (.+)$`)
	cleanupPruneRE   = regexp.MustCompile(`^Would prune ([\d,]+) files? from: (.+)$`)
	cleanupAbvRE     = regexp.MustCompile(`^(?:([\d,]+) files?, )?(\S+)$`)
	autoremoveHeadRE = regexp.MustCompile(`^Would autoremove (\d+) unneeded formulae?:$`)
	formulaNameRE    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9@+._/-]*$`)
	sizeRE           = regexp.MustCompile(`^([\d.]+)(B|KB|MB|GB)$`)
)

// sizeUnits are brew's units: disk sizes are in powers of 1000 (Formatter).
var sizeUnits = map[string]float64{"B": 1, "KB": 1e3, "MB": 1e6, "GB": 1e9}

// CleanupPreview is `brew cleanup -n`: what cleanup would remove. It is a read,
// not a write, so it does not go through the task manager and changes nothing.
func (s *diagnosticsReader) CleanupPreview(ctx context.Context) (CleanupPreview, error) {
	output, err := s.runner.Execute(ctx, "cleanup", "-n")
	text, err := previewOutput(output, err)
	if err != nil {
		return CleanupPreview{}, err
	}
	return parseCleanupPreview(text), nil
}

// AutoremovePreview is `brew autoremove -n`: the formulae it would uninstall.
// brew prints no heading when none are unused, so that is an empty preview.
func (s *diagnosticsReader) AutoremovePreview(ctx context.Context) (AutoremovePreview, error) {
	output, err := s.runner.Execute(ctx, "autoremove", "-n")
	text, err := previewOutput(output, err)
	if err != nil {
		return AutoremovePreview{}, err
	}
	return parseAutoremovePreview(text), nil
}

// previewOutput is a dry run's output, or its error. Exit 1 with a listing is a
// result, as it is for doctor and missing: brew exits 1 when a keg cannot be
// removed and still lists what it would do.
func previewOutput(output []byte, err error) (string, error) {
	if err == nil || (IsExitCode(err, 1) && strings.TrimSpace(string(output)) != "") {
		return string(output), nil
	}
	return "", err
}

// previewLines is the non-blank lines of a dry run, with brew's "==> " headline
// marker removed, so a heading and a plain line read the same.
func previewLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "==> ")); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func parseCleanupPreview(text string) CleanupPreview {
	var p CleanupPreview
	lines := previewLines(text)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := cleanupTotalRE.FindStringSubmatch(line); m != nil {
			p.Total = m[1]
		} else if n, ok := autoremoveCount(line); ok {
			var names []string
			names, i = formulaeAfter(lines, i, n)
			p.Unused = append(p.Unused, names...)
		} else if item, ok := parseCleanupItem(line); ok {
			p.Items = append(p.Items, item)
		} else {
			p.Raw = append(p.Raw, line)
		}
	}
	return p
}

func parseAutoremovePreview(text string) AutoremovePreview {
	var p AutoremovePreview
	lines := previewLines(text)
	for i := 0; i < len(lines); i++ {
		if n, ok := autoremoveCount(lines[i]); ok {
			var names []string
			names, i = formulaeAfter(lines, i, n)
			p.Formulae = append(p.Formulae, names...)
		} else {
			p.Raw = append(p.Raw, lines[i])
		}
	}
	return p
}

// autoremoveCount reads the heading brew prints before the names an autoremove
// would uninstall, and the number of names it says follow.
func autoremoveCount(line string) (int, bool) {
	m := autoremoveHeadRE.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

// formulaeAfter reads up to n formula names from the lines after lines[i]. It
// stops at the first line that is not a name, so a later "Would remove" line is
// never taken as one. It returns the names and the index of the last line read.
func formulaeAfter(lines []string, i, n int) ([]string, int) {
	var names []string
	for ; n > 0 && i+1 < len(lines) && formulaNameRE.MatchString(lines[i+1]); n-- {
		i++
		names = append(names, lines[i])
	}
	return names, i
}

// parseCleanupItem reads one path line: a removal with its size, a removal with
// a label instead of a size, or a prune of files out of a directory.
func parseCleanupItem(line string) (CleanupItem, bool) {
	if m := cleanupKindRE.FindStringSubmatch(line); m != nil {
		return CleanupItem{Path: m[2], Kind: m[1]}, true
	}
	if m := cleanupPruneRE.FindStringSubmatch(line); m != nil {
		return CleanupItem{Path: m[2], Files: atoiComma(m[1])}, true
	}
	m := cleanupRemoveRE.FindStringSubmatch(line)
	if m == nil {
		return CleanupItem{}, false
	}
	item := CleanupItem{Path: m[1], Size: m[2]}
	if abv := cleanupAbvRE.FindStringSubmatch(m[2]); abv != nil {
		item.Size = abv[2]
		item.Files = 1
		if abv[1] != "" {
			item.Files = atoiComma(abv[1])
		}
	}
	item.Bytes, _ = parseSize(item.Size)
	return item, true
}

// parseSize reads brew's size text ("230.3MB") as bytes. It reports false when
// the text is not a size.
func parseSize(s string) (int64, bool) {
	m := sizeRE.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return int64(math.Round(n * sizeUnits[m[2]])), true
}

// atoiComma reads a count brew writes with thousands separators ("1,234").
func atoiComma(s string) int {
	n, _ := strconv.Atoi(strings.ReplaceAll(s, ",", ""))
	return n
}
