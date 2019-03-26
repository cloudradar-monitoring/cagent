package xrandr

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// RefreshRateValue refresh rate value
type RefreshRateValue float32

// Size screen size or resolution
type Size struct {
	Width  float32
	Height float32
}

// Position screen position
type Position struct {
	X int
	Y int
}

// RefreshRate mode refresh rate
type RefreshRate struct {
	Value     RefreshRateValue
	Current   bool
	Preferred bool
}

// Mode xrandr output mode
type Mode struct {
	Resolution   Size
	RefreshRates []RefreshRate
}

// Monitor all the info of xrandr output
type Monitor struct {
	ID         string
	Modes      []Mode
	Primary    bool
	Size       Size
	Connected  bool
	Resolution Size
	Position   Position
}

// Screen all the info of xrandr screen
type Screen struct {
	No                int
	CurrentResolution Size
	MinResolution     Size
	MaxResolution     Size
	Monitors          []Monitor
}

// Screens slice of screens
type Screens []Screen

// MonitorByID returns the output with the given ID
func (screens Screens) MonitorByID(ID string) (Monitor, bool) {
	m := Monitor{}

	for _, screen := range screens {
		m, ok := screen.MonitorByID(ID)
		if ok {
			return m, true
		}
	}

	return m, false
}

// MonitorByID returns the output with the given ID
func (s Screen) MonitorByID(ID string) (Monitor, bool) {
	for _, monitor := range s.Monitors {
		if monitor.ID == ID {
			return monitor, true
		}
	}

	return Monitor{}, false
}

// CurrentRefreshRate returns mode current refresh rate
func (md Mode) CurrentRefreshRate() (RefreshRate, bool) {
	for _, r := range md.RefreshRates {
		if r.Current {
			return r, true
		}
	}

	return RefreshRate{}, false
}

// CurrentMode returns monitor/output current mode
func (m Monitor) CurrentMode() (Mode, bool) {
	for _, mode := range m.Modes {
		if _, ok := mode.CurrentRefreshRate(); ok {
			return mode, true
		}
	}

	return Mode{}, false
}

// SizeMM returns monitor/output size in milimeters
func (m Monitor) SizeMM() Size {
	return m.Size
}

// SizeIn returns monitor/output size in inches
func (m Monitor) SizeIn() Size {
	var mmToInch float32 = 0.0393701

	sizeIn := Size{
		Width:  m.Size.Width * mmToInch,
		Height: m.Size.Height * mmToInch,
	}

	return sizeIn
}

// DPI returns monitor/output dpi
func (m Monitor) DPI() (float32, error) {
	hDPI, err := m.VerticalDPI()
	if err != nil {
		return 0, err
	}

	vDPI, err := m.HorizontalDPI()
	if err != nil {
		return 0, err
	}

	return float32(math.Min(float64(hDPI), float64(vDPI))), nil
}

// HorizontalDPI monitor horizontal dpi
func (m Monitor) HorizontalDPI() (float32, error) {
	cm, ok := m.CurrentMode()
	if !ok {
		return 0, fmt.Errorf(`cannot determine monitor current mode: %s`, m.ID)
	}

	size := m.SizeIn()

	return float32(cm.Resolution.Width) / float32(size.Width), nil
}

// VerticalDPI monitor vertical dpi
func (m Monitor) VerticalDPI() (float32, error) {
	cm, ok := m.CurrentMode()
	if !ok {
		return 0, fmt.Errorf(`cannot determine monitor current mode: %s`, m.ID)
	}

	size := m.SizeIn()

	return float32(cm.Resolution.Height) / float32(size.Height), nil
}

// Rescale recalculate monitor/output resolution
func (s Size) Rescale(scale float32) Size {
	s.Width = float32(math.Ceil(float64(s.Width) * float64(scale)))
	s.Height = float32(math.Ceil(float64(s.Height) * float64(scale)))

	return s
}

// GetScreens returns all the screens info from xrandr output
func GetScreens() (Screens, error) {
	_, err := exec.LookPath("xrandr")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("xrandr")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	screens, err := parseScreens(string(output))
	if err != nil {
		return nil, err
	}

	return screens, nil
}

func parseScreenLine(line string) (*Screen, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "Screen") {
		return nil, fmt.Errorf("invalid screen line: %s", line)
	}

	re := regexp.MustCompile(`Screen \d+`)
	screenStr := re.FindString(line)
	if screenStr == "" {
		return nil, fmt.Errorf("unexpected screen line format: %s", line)
	}
	no, err := strconv.Atoi(strings.Split(screenStr, " ")[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing screen number: %s", err)
	}

	screen := &Screen{
		No: no,
	}

	parseScreenResolution := func(s, typ string) (*Size, error) {
		if !strings.HasPrefix(s, typ) {
			return nil, fmt.Errorf("expected to start with %s", typ)
		}
		s = strings.Replace(s, typ, "", -1)
		s = strings.TrimSpace(s)
		resolution, err := parseSize(s)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s screen resolution: %s", typ, err)
		}

		return resolution, nil
	}

	for _, typ := range []string{"minimum", "current", "maximum"} {
		re = regexp.MustCompile(fmt.Sprintf(`%s \d+\s*x\s*\d+`, typ))
		resStr := re.FindString(line)
		if resStr == "" {
			return nil, fmt.Errorf("%s resolution could not be found: %s", typ, line)
		}

		resolution, err := parseScreenResolution(resStr, typ)
		if err != nil {
			return nil, err
		}

		switch typ {
		case "minimum":
			screen.MinResolution = *resolution
		case "current":
			screen.CurrentResolution = *resolution
		case "maximum":
			screen.MaxResolution = *resolution
		}
	}

	return screen, nil
}

func parseMonitorLine(line string) (*Monitor, error) {
	line = strings.TrimSpace(line)
	tokens := strings.SplitN(line, " ", 2)
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid monitor line format: %s", line)
	}

	id := tokens[0]
	monitor := Monitor{
		ID: id,
	}

	connected := strings.Contains(line, " connected ")
	if !connected {
		return &monitor, nil
	}

	primary := strings.Contains(line, "primary")
	re := regexp.MustCompile(`\d+mm\s*x\s*\d+mm`)
	sizeStr := re.FindString(tokens[1])
	if sizeStr == "" {
		return nil, fmt.Errorf("could not determine monitor size (mm), expected WWWmm x HHHmm: %s", line)
	}

	sizeStr = strings.Replace(sizeStr, "mm", "", 2)
	size, err := parseSize(sizeStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing monitor size: %s", err)
	}

	re = regexp.MustCompile(`\d+\s*x\s*\d+\+\d+\+\d+`)
	resStr := re.FindString(tokens[1])
	if resStr == "" {
		return nil, fmt.Errorf("could not determine monitor resolution and position, expected WxH+X+Y: %s", line)
	}

	resolution, position, err := parseSizeWithPosition(resStr)
	if err != nil {
		return nil, fmt.Errorf("could not determine monitor resolution and position: %s", err)
	}

	monitor.Connected = true
	monitor.Primary = primary
	monitor.Size = *size
	monitor.Position = *position
	monitor.Resolution = *resolution

	return &monitor, nil
}

func parseModeLine(line string) (*Mode, error) {
	line = strings.TrimSpace(line)
	mode := Mode{}

	ws := bufio.NewScanner(strings.NewReader(line))
	ws.Split(bufio.ScanWords)
	for ws.Scan() {
		w := ws.Text()
		if strings.Contains(w, "x") {
			res, err := parseSize(w)
			if err != nil {
				return nil, err
			}

			mode.Resolution = *res
			continue
		}

		rate, err := parseRefreshRate(w)
		if err != nil {
			return nil, err
		}

		mode.RefreshRates = append(mode.RefreshRates, *rate)
	}

	return &mode, nil
}

func parseSize(s string) (*Size, error) {
	if !strings.Contains(s, "x") {
		return nil, fmt.Errorf("invalid size format; expected format WxH but got %s", s)
	}

	res := strings.Split(s, "x")
	width, err := strconv.Atoi(strings.TrimSpace(res[0]))
	if err != nil {
		return nil, fmt.Errorf("could not parse mode width size (%s): %s", s, err)
	}

	height, err := strconv.Atoi(strings.TrimSpace(res[1]))
	if err != nil {
		return nil, fmt.Errorf("could not parse mode height size (%s): %s", s, err)
	}

	return &Size{
		Width:  float32(width),
		Height: float32(height),
	}, nil
}

func parseSizeWithPosition(s string) (*Size, *Position, error) {
	tokens := strings.SplitN(s, "+", 2)
	size, err := parseSize(tokens[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid resolution with position format; expected WxH+X+Y, got %s: %s", s, err)
	}

	tokens = strings.Split(tokens[1], "+")
	if len(tokens) != 2 {
		return nil, nil, fmt.Errorf("invalid position format; expected X+Y, got %s", tokens)
	}

	x, err := strconv.Atoi(strings.TrimSpace(tokens[0]))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid position X: %s", err)
	}

	y, err := strconv.Atoi(strings.TrimSpace(tokens[1]))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid position Y: %s", err)
	}

	position := Position{
		X: x,
		Y: y,
	}

	return size, &position, nil
}

func parseRefreshRate(s string) (*RefreshRate, error) {
	s = strings.TrimSpace(s)
	current := strings.Contains(s, "*")
	preferred := strings.Contains(s, "+")

	s = strings.TrimSpace(strings.Trim(s, "*+ "))
	value, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid rate value (%s): %s", s, err)
	}

	return &RefreshRate{
		Value:     RefreshRateValue(value),
		Current:   current,
		Preferred: preferred,
	}, nil
}

func isScreenLine(l string) bool {
	return strings.HasPrefix(l, "Screen")
}

func isMonitorLine(l string) bool {
	return strings.Contains(l, "connected") || strings.Contains(l, "disconnected")
}

func parseScreens(s string) ([]Screen, error) {
	var screens []Screen

	ls := bufio.NewScanner(strings.NewReader(s))
	ls.Split(bufio.ScanLines)

	var err error
	var screen *Screen
	var monitor *Monitor
	var mode *Mode

	for ls.Scan() {
		l := ls.Text()
		if isScreenLine(l) {
			if screen != nil {
				screens = append(screens, *screen)
			}

			screen, err = parseScreenLine(l)
			if err != nil {
				return nil, err
			}

			continue
		}

		if isMonitorLine(l) {
			if monitor != nil {
				screen.Monitors = append(screen.Monitors, *monitor)
			}

			monitor, err = parseMonitorLine(l)
			if err != nil {
				return nil, err
			}

			continue
		}

		if monitor != nil && !monitor.Connected {
			continue
		}

		mode, err = parseModeLine(l)
		if err != nil {
			return nil, err
		}

		monitor.Modes = append(monitor.Modes, *mode)
	}

	screen.Monitors = append(screen.Monitors, *monitor)
	screens = append(screens, *screen)

	return screens, nil
}
