package gui

type batchState struct {
	selected map[int]bool
}

func newBatchState() *batchState {
	return &batchState{selected: make(map[int]bool)}
}

// toggle flips the selection for idx. A de-selection removes the key
// entirely: the map holds only actually-selected indexes, so stale false
// entries can never fool iteration, count or select-all checks
// (r1 finding 1).
func (b *batchState) toggle(idx int) {
	if b.selected[idx] {
		delete(b.selected, idx)
	} else {
		b.selected[idx] = true
	}
}

type mutationType int

const (
	mutInstall mutationType = iota
	mutUninstall
	mutReinstall
	mutUpgrade
	mutUpgradeAll
	mutZap
	mutFetch
)

// mutationID builds a unique-per-op task ID. Bare package names are not
// unique across concurrent tasks (a batch upgrade of "foo" plus a manual
// "foo" reinstall used to share an ID, corrupting the batch counter's
// membership map — r2 finding NEW-3), so every mutation is prefixed with
// its op type. Batch upgrades use the dedicated "batch-upgrade:" prefix
// so they can never collide with a manual mutation of the same package.
func mutationID(t mutationType, name string) string {
	switch t {
	case mutInstall:
		return "install:" + name
	case mutUninstall:
		return "uninstall:" + name
	case mutReinstall:
		return "reinstall:" + name
	case mutUpgrade:
		return "upgrade:" + name
	case mutUpgradeAll:
		return "upgrade-all"
	case mutZap:
		return "zap:" + name
	case mutFetch:
		return "fetch:" + name
	default:
		return "mutation:" + name
	}
}

type MutationResultMsg struct {
	Err    error
	Name   string
	Type   mutationType
	Leaves []string
}

type ProgressLineMsg struct {
	Line string
}

type opStatus int

const (
	opRunning opStatus = iota
	opSuccess
	opError
	opCancelled
)

type Operation struct {
	Title  string
	Lines  []string
	Status opStatus
	Err    error
}

func (o *Operation) AppendLine(line string) {
	o.Lines = append(o.Lines, line)
}

func (o *Operation) Running() bool {
	return o.Status == opRunning
}

func (o *Operation) Done() bool {
	return o.Status != opRunning
}
