package gui

type batchState struct {
	selected map[int]bool
}

func newBatchState() *batchState {
	return &batchState{selected: make(map[int]bool)}
}

func (b *batchState) toggle(idx int) {
	b.selected[idx] = !b.selected[idx]
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

type MutationResultMsg struct {
	Err    error
	Name   string
	Type   mutationType
	Leaves []string
}

type ProgressLineMsg struct {
	Line string
}

type ProgressCompleteMsg struct {
	Err  error
	Name string
}
