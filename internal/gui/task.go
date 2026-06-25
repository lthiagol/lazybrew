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
