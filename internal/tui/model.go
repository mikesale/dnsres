package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"dnsres/internal/dnsres"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resolverEventMsg dnsres.ResolverEvent

type healthTickMsg struct{}

type resolverErrMsg struct {
	err error
}

type serverState struct {
	lastHostname string
	lastLatency  time.Duration
	lastError    string
	lastSuccess  time.Time
	lastFailure  time.Time
	total        int
	failures     int
	lastSource   string
}

type model struct {
	config       *dnsres.Config
	resolver     *dnsres.DNSResolver
	cancel       context.CancelFunc
	events       <-chan dnsres.ResolverEvent
	unsubscribe  func()
	resolverErrs <-chan error
	spinner      spinner.Model
	table        table.Model
	viewport     viewport.Model
	activity     []string
	servers      map[string]*serverState
	serverOrder  []string
	health       map[string]bool
	cycleRunning bool
	cycleStart   time.Time
	lastCycle    time.Time
	lastCycleDur time.Duration
	width        int
	height       int
	ready        bool
	statusMsg    string
}

func newModel(resolver *dnsres.DNSResolver, config *dnsres.Config, cancel context.CancelFunc, events <-chan dnsres.ResolverEvent, unsubscribe func(), errs <-chan error) *model {
	spin := spinner.New()
	spin.Spinner = spinner.Dot

	columns := []table.Column{
		{Title: "Server", Width: 18},
		{Title: "Health", Width: 8},
		{Title: "Last OK", Width: 9},
		{Title: "Latency", Width: 10},
		{Title: "Last Error", Width: 32},
	}
	rows := []table.Row{}
	tableModel := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.BorderStyle(asciiBorder).BorderBottom(true).Bold(true)
	styles.Selected = styles.Selected.Foreground(lipgloss.Color("")).Background(lipgloss.Color(""))
	tableModel.SetStyles(styles)

	vp := viewport.New(0, 0)
	serverOrder := make([]string, 0, len(config.DNSServers))
	servers := make(map[string]*serverState, len(config.DNSServers))
	for _, server := range config.DNSServers {
		serverOrder = append(serverOrder, server)
		servers[server] = &serverState{}
	}

	m := &model{
		config:       config,
		resolver:     resolver,
		cancel:       cancel,
		events:       events,
		unsubscribe:  unsubscribe,
		resolverErrs: errs,
		spinner:      spin,
		table:        tableModel,
		viewport:     vp,
		activity:     []string{},
		servers:      servers,
		serverOrder:  serverOrder,
		health:       map[string]bool{},
	}

	// Check for log directory fallback
	if resolver.LogDirWasFallback() {
		m.statusMsg = fmt.Sprintf("Logs: %s (fallback)", resolver.GetLogDir())
	}

	m.updateTableRows()
	return m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.events), tickHealth(), m.spinner.Tick, waitForResolverErr(m.resolverErrs))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "q", "ctrl+c":
			if m.cancel != nil {
				m.cancel()
			}
			if m.unsubscribe != nil {
				m.unsubscribe()
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.ready = true
		m.resize()
	case resolverEventMsg:
		m.applyEvent(dnsres.ResolverEvent(typed))
		m.updateTableRows()
		return m, waitForEvent(m.events)
	case healthTickMsg:
		m.health = m.resolver.HealthSnapshot()
		m.updateTableRows()
		return m, tickHealth()
	case resolverErrMsg:
		if typed.err != nil {
			m.appendActivity(fmt.Sprintf("resolver error: %v", typed.err))
		}
		if m.cancel != nil {
			m.cancel()
		}
		if m.unsubscribe != nil {
			m.unsubscribe()
		}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	summaryWidth := clamp(m.width/3, 32, 48)
	tableWidth := max(m.width-summaryWidth-1, 20)

	summary := panelStyle.Width(summaryWidth).Render(m.summaryView())
	tablePanel := panelStyle.Width(tableWidth).Render(m.table.View())
	top := lipgloss.JoinHorizontal(lipgloss.Top, summary, tablePanel)

	activityPanel := panelStyle.Width(m.width).Render(m.viewport.View())
	return lipgloss.JoinVertical(lipgloss.Left, top, activityPanel)
}

func (m *model) summaryView() string {
	status := "idle"
	if m.cycleRunning {
		status = fmt.Sprintf("%s running", m.spinner.View())
	}

	lastCycle := "n/a"
	if m.lastCycleDur > 0 {
		lastCycle = m.lastCycleDur.Round(time.Millisecond).String()
	}

	lastCompleted := "n/a"
	if !m.lastCycle.IsZero() {
		lastCompleted = m.lastCycle.Format("15:04:05")
	}

	healthyCount, unhealthyCount := m.healthCounts()
	lines := []string{
		titleStyle.Render("dnsres TUI"),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Hostnames: %d", len(m.config.Hostnames)),
		fmt.Sprintf("Servers: %d", len(m.config.DNSServers)),
		fmt.Sprintf("Interval: %s", m.config.QueryInterval.Duration),
		fmt.Sprintf("Last cycle: %s", lastCycle),
		fmt.Sprintf("Last done: %s", lastCompleted),
		fmt.Sprintf("Health: %s / %s", goodStyle.Render(fmt.Sprintf("%d up", healthyCount)), badStyle.Render(fmt.Sprintf("%d down", unhealthyCount))),
	}

	if m.statusMsg != "" {
		lines = append(lines, warnStyle.Render(m.statusMsg))
	}

	lines = append(lines, mutedStyle.Render("q to quit"))
	return strings.Join(lines, "\n")
}

func (m *model) healthCounts() (int, int) {
	healthy := 0
	unhealthy := 0
	for _, server := range m.serverOrder {
		status, ok := m.health[server]
		if !ok {
			continue
		}
		if status {
			healthy++
		} else {
			unhealthy++
		}
	}
	return healthy, unhealthy
}

func (m *model) resize() {
	if m.width == 0 || m.height == 0 {
		return
	}

	topHeight := clamp(len(m.serverOrder)+4, 8, m.height-6)
	activityHeight := max(m.height-topHeight-2, 3)

	innerWidth := max(m.width-2, 10)
	innerActivityWidth := max(innerWidth, 10)
	innerTableWidth := max((m.width/2)-2, 20)

	m.table.SetHeight(max(topHeight-3, 3))
	m.setTableColumns(innerTableWidth)

	m.viewport.Width = innerActivityWidth
	m.viewport.Height = activityHeight
}

func (m *model) setTableColumns(width int) {
	serverWidth := clamp(width/4, 12, 24)
	healthWidth := 8
	lastOKWidth := 9
	latencyWidth := 10
	remaining := width - (serverWidth + healthWidth + lastOKWidth + latencyWidth + 4)
	if remaining < 12 {
		remaining = 12
	}

	m.table.SetColumns([]table.Column{
		{Title: "Server", Width: serverWidth},
		{Title: "Health", Width: healthWidth},
		{Title: "Last OK", Width: lastOKWidth},
		{Title: "Latency", Width: latencyWidth},
		{Title: "Last Error", Width: remaining},
	})
}

func (m *model) applyEvent(event dnsres.ResolverEvent) {
	switch event.Type {
	case dnsres.EventCycleStart:
		m.cycleRunning = true
		m.cycleStart = event.Time
		m.appendActivity(fmt.Sprintf("cycle start hostnames=%d servers=%d", event.HostnameCount, event.ServerCount))
	case dnsres.EventCycleComplete:
		m.cycleRunning = false
		m.lastCycleDur = event.Duration
		m.lastCycle = event.Time
		m.appendActivity(fmt.Sprintf("cycle complete duration=%s", event.Duration.Round(time.Millisecond)))
	case dnsres.EventResolveSuccess:
		state := m.ensureServer(event.Server)
		state.lastHostname = event.Hostname
		state.lastLatency = event.Duration
		state.lastError = ""
		state.lastSuccess = event.Time
		state.total++
		state.lastSource = event.Source
		m.appendActivity(fmt.Sprintf("resolved %s via %s (%s)", event.Hostname, event.Server, formatDuration(event.Duration, event.Source)))
	case dnsres.EventResolveFailure:
		state := m.ensureServer(event.Server)
		state.lastHostname = event.Hostname
		state.lastLatency = event.Duration
		state.lastError = event.Error
		state.lastFailure = event.Time
		state.failures++
		state.lastSource = event.Source
		m.appendActivity(fmt.Sprintf("failed %s via %s (%s)", event.Hostname, event.Server, formatFailure(event)))
	case dnsres.EventInconsistent:
		m.appendActivity(fmt.Sprintf("inconsistent responses for %s", event.Hostname))
	}
}

func (m *model) ensureServer(server string) *serverState {
	state, ok := m.servers[server]
	if !ok {
		state = &serverState{}
		m.servers[server] = state
		m.serverOrder = append(m.serverOrder, server)
		m.serverOrder = uniqueSorted(m.serverOrder)
	}
	return state
}

func (m *model) updateTableRows() {
	rows := make([]table.Row, 0, len(m.serverOrder))
	for _, server := range m.serverOrder {
		state := m.servers[server]
		if state == nil {
			state = &serverState{}
		}
		healthValue := "?"
		if status, ok := m.health[server]; ok {
			if status {
				healthValue = goodStyle.Render("up")
			} else {
				healthValue = badStyle.Render("down")
			}
		} else {
			healthValue = warnStyle.Render("unknown")
		}

		lastOK := "-"
		if !state.lastSuccess.IsZero() {
			lastOK = state.lastSuccess.Format("15:04:05")
		}
		latency := "-"
		if state.lastLatency > 0 {
			latency = state.lastLatency.Round(time.Millisecond).String()
		}
		lastErr := "-"
		if state.lastError != "" {
			lastErr = state.lastError
		}

		rows = append(rows, table.Row{server, healthValue, lastOK, latency, lastErr})
	}
	m.table.SetRows(rows)
}

func (m *model) appendActivity(entry string) {
	stamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("%s %s", stamp, entry)
	m.activity = append(m.activity, line)
	if len(m.activity) > 200 {
		m.activity = m.activity[len(m.activity)-200:]
	}
	m.viewport.SetContent(strings.Join(m.activity, "\n"))
	m.viewport.GotoBottom()
}

func waitForEvent(events <-chan dnsres.ResolverEvent) tea.Cmd {
	return func() tea.Msg {
		if events == nil {
			return nil
		}
		event, ok := <-events
		if !ok {
			return nil
		}
		return resolverEventMsg(event)
	}
}

func waitForResolverErr(errs <-chan error) tea.Cmd {
	return func() tea.Msg {
		if errs == nil {
			return nil
		}
		err, ok := <-errs
		if !ok {
			return nil
		}
		return resolverErrMsg{err: err}
	}
}

func tickHealth() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return healthTickMsg{}
	})
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func formatDuration(duration time.Duration, source string) string {
	if source == "cache" {
		return "cache hit"
	}
	if duration == 0 {
		return source
	}
	return duration.Round(time.Millisecond).String()
}

func formatFailure(event dnsres.ResolverEvent) string {
	if event.Source != "" {
		return fmt.Sprintf("%s: %s", event.Source, event.Error)
	}
	return event.Error
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
