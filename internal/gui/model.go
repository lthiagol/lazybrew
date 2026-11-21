package gui

import (
	"os/exec"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/task"
	"github.com/lthiagol/lazybrew/internal/logtail"
)

type Model struct {
	client           *brew.Client
	cfg              *config.Config
	width            int
	height           int
	terminalTooSmall bool
	activePanel      PanelID
	panels           []*panelData
	activeTab        int
	tabs             []tabInfo
	ready            bool
	// contentScroll is how far the wheel has scrolled each tab item's content,
	// in rows, keyed by contentKey. An item starts at the top.
	contentScroll map[string]int
	// wrap holds the wrapped rows of the Files and Used By tabs, so scrolling
	// them does not render them again (cr3 AC-04).
	wrap *wrapCache

	activeModal modal.Modal
	// toasts is the stack, oldest first; the newest is drawn lowest.
	toasts []*modal.Toast
	// toastTicking is set while the dismissal tick is in flight, so a stack
	// has one tick chain, not one per toast.
	toastTicking bool
	batch        *batchState
	showHelp     bool
	// helpScroll is how far the help text is scrolled, in rows. It starts at
	// the top each time help opens.
	helpScroll     int
	tabContent     map[string]string
	pendingAction  string
	pendingService string
	pendingMutType mutationType
	// pendingLabel is the verb the confirmed mutation's task is titled with.
	pendingLabel string
	batchCount   int
	// batchIDs counts the occurrences of each task ID enqueued by the
	// current batch upgrade, so batchCount is only decremented by batch
	// completions, never by unrelated ones (r1 finding 27). A formula and
	// a cask can share a name in one batch, yielding two tasks with the
	// same ID: counting occurrences (instead of a bool set) keeps the
	// counter honest until the last duplicate completes (r3 finding R3-2).
	batchIDs map[string]int
	// batchTotal is the enqueued size of the current batch; together with
	// batchCount it surfaces "n/total done" in the Operation overlay so
	// the counter machinery has a user-visible effect (r3 finding R3-5).
	batchTotal int
	// batchSeq numbers each batch so its task IDs
	// ("batch-upgrade:<seq>:<name>") cannot collide with a task still
	// queued from a superseded batch (r4 finding R4-4).
	batchSeq int
	// pendingTrustFormulaIdx / pendingTrustCaskIdx record the positions
	// of the submenu entries in the trust menu built by startTrustMenu;
	// they exist only when the tap has formulae/casks, so their
	// positions vary and must not be hardcoded (r4 finding R4-3).
	pendingTrustFormulaIdx int
	pendingTrustCaskIdx    int
	lastUpdate             time.Time
	isUpdating             bool
	// updateChecked is set once auto mode's launch read (lastUpdateMsg) lands.
	updateChecked bool
	// launchStale is set when that read found brew stale and ran brew update.
	launchStale bool
	// updateRefreshed is set when a brew-update completion has issued its
	// load-flagged refresh. A launch read that lands after it is dropped.
	updateRefreshed bool
	updateOutput    []string

	// logs is the Services Logs tab's view of its service's log (logs.go).
	logs logsView
	// readLog reads the tail of a log file. New sets logtail.Read; a test swaps
	// it in to count the reads.
	readLog func(path string, n int) (logtail.Tail, error)

	searchResults []brew.SearchResult
	// remarkGen numbers the Search re-marks issued. Only the latest one may land:
	// an older one read the installed lists before a later write (LB-20 b).
	remarkGen int
	// searchInfo is the info of the selected search result, parsed once when
	// it lands. searchInfoErr is why there is none instead (cr3 AC-01).
	searchInfo    *pkgInfo
	searchInfoErr error
	searchInput   textinput.Model

	tasks      *task.Manager
	commandLog *CommandLog
	spinner    spinner.Model
	opState    *Operation
	// execProcess runs a command with the TUI suspended and the terminal
	// attached (tea.ExecProcess). New sets it; a test swaps it in to capture
	// the command (terminal_retry.go).
	execProcess func(*exec.Cmd, tea.ExecCallback) tea.Cmd
	// setClipboard copies text to the terminal clipboard (clipboard.go). New
	// sets tea.SetClipboard; a test swaps it in to read what would be copied.
	setClipboard func(string) tea.Cmd
	// taskID was removed (r3 finding R3-6): nothing read it, and stale
	// output filtering belongs to the TaskManager, not the model.
	refreshing int

	// logCollapsed hides the command log pane (Shift+C / "C"). Default false.
	logCollapsed bool

	// icons is the glyph set every renderer draws with; resolved once in New.
	icons style.Icons
	// status is what the last dashboard load read: the header's version and
	// counts, and the Status sidebar's rows. Empty until the dashboard loads.
	status statusSummary
	// preload is the background load of the panels after the Status dashboard
	// (gui.preload); see preload.go.
	preload preloadState
	// landed is the counts the panel landings have read with the preload on. A
	// Status load fills its unknown counts from it (fillUnknown).
	landed statusSummary

	// lastKeyAt tracks the last time the user pressed a key. The auto
	// refresh tick consults this to pause refreshes during interaction
	// (M12 AC-03) so typing/navigation is not janked by a full
	// refresh storm.
	lastKeyAt time.Time
	// autoRefreshPaused is set by the tick handler when lastKeyAt is
	// within the pause window. The next tick will re-evaluate; the
	// pending refresh is rescheduled for one AutoRefreshSeconds later.
	autoRefreshPaused bool
	// autoRefreshTicking is set while an auto-refresh tick is in flight, so the
	// chain has one tick at a time (armAutoRefresh). updateTicking does the same
	// for the update ticker.
	autoRefreshTicking bool
	updateTicking      bool

	// motion holds the values the screen animates, and motionOn is whether it
	// does (gui.animations, LAZYBREW_ANIMATIONS). motionTicking is set while a
	// frame tick is in flight, so a screen has one tick chain; motionAt is the
	// time of the last frame, zero when no run is under way.
	motion        *anim.Set
	motionOn      bool
	motionTicking bool
	motionAt      time.Time
	// motionPanel is the panel the tab underline last aimed at, so a switch
	// to another panel snaps the underline rather than sliding it.
	motionPanel PanelID
	// motionSize is the window size the animated values last aimed at. A
	// resize snaps them rather than sliding them.
	motionSize [2]int
	// motionShown is the modal whose entrance was last started, so a new modal
	// slides in once and an open one does not restart.
	motionShown modal.Modal
	// motionStack is the toast stack the slots were last laid out for.
	motionStack []*modal.Toast
	// toastMotionNames are the motion names of the toasts in that stack, by
	// toast id, so a frame reads their offsets without building names (cr3
	// F-r1-02).
	toastMotionNames map[uint64]toastNames
	// opBar is the batch's progress bar, and opBarOp the operation it belongs
	// to. Its fill springs to the batch's share when the count changes.
	opBar   progress.Model
	opBarOp *Operation
}

func (m *Model) Cfg() *config.Config { return m.cfg }

func New(client *brew.Client, cfg *config.Config) *Model {
	panels := initPanels()
	cl := NewCommandLog(20)
	s := spinner.New(spinner.WithStyle(style.AccentText))
	s.Spinner = spinner.Dot
	ti := textinput.New()
	ti.Placeholder = "Search packages..."
	ti.SetWidth(40)
	ti.SetStyles(searchStyles())
	icons := iconSet(cfg)
	cl.SetIcons(icons)
	m := &Model{
		client:        client,
		cfg:           cfg,
		activePanel:   PanelStatus,
		panels:        panels,
		activeTab:     0,
		tabs:          panelTabs[PanelStatus],
		batch:         newBatchState(),
		tabContent:    make(map[string]string),
		tasks:         task.NewManager(task.DefaultMaxQueue),
		commandLog:    cl,
		spinner:       s,
		searchInput:   ti,
		icons:         icons,
		contentScroll: make(map[string]int),
		wrap:          newWrapCache(),
		motion:        anim.NewSet(),
		motionOn:      animationsOn(cfg),
		opBar:         newProgressBar(),
		readLog:       logtail.Read,
		execProcess:   tea.ExecProcess,
		setClipboard:  tea.SetClipboard,
	}
	// -1 marks "no such menu entry" until startTrustMenu records the
	// real positions (r4 finding R4-3).
	m.pendingTrustFormulaIdx = -1
	m.pendingTrustCaskIdx = -1
	return m
}

func (m *Model) CommandLogCallback() brew.CommandCallback {
	cl := m.commandLog
	return func(args []string, err error) {
		id := cl.Append(brewCommandString(args))
		if err != nil {
			cl.SetStatus(id, CommandError)
		} else {
			cl.SetStatus(id, CommandSuccess)
		}
	}
}

func (m *Model) CommandLogStartCallback() brew.CommandCallback {
	cl := m.commandLog
	return func(args []string, _ error) {
		cl.Append(brewCommandString(args))
	}
}
