package xrandr

import (
	"fmt"
	"os/exec"
	"strconv"
)

// CommandBuilder xrandr command builder
type CommandBuilder struct {
	dpi                   int
	screenSize            Size
	outputCommandBuilders []OutputCommandBuilder
}

// OutputCommandBuilder xrandr command builder for flag --output
type OutputCommandBuilder struct {
	monitor Monitor
	scale   float32
	leftOf  *Monitor
	primary bool
	parent  CommandBuilder
}

// DPI sets the dpi --dpi flag
func (cb CommandBuilder) DPI(dpi int) CommandBuilder {
	cb.dpi = dpi

	return cb
}

// ScreenSize sets the screen size --fb flag
func (cb CommandBuilder) ScreenSize(size Size) CommandBuilder {
	cb.screenSize = size

	return cb
}

// Output starts a new output command
func (cb CommandBuilder) Output(monitor Monitor) OutputCommandBuilder {
	ocb := NewOutputCommand(monitor)
	ocb.parent = cb
	cb.AddOutputCommand(ocb)

	return ocb
}

// AddOutputCommand adds a new output command to command builder
func (cb CommandBuilder) AddOutputCommand(ocb OutputCommandBuilder) CommandBuilder {
	cb.outputCommandBuilders = append(cb.outputCommandBuilders, ocb)

	return cb
}

// NewOutputCommand output command builder constructor
func NewOutputCommand(monitor Monitor) OutputCommandBuilder {
	return OutputCommandBuilder{monitor: monitor}
}

// Scale sets the --scale flag
func (ocb OutputCommandBuilder) Scale(scale float32) OutputCommandBuilder {
	ocb.scale = scale

	return ocb
}

// Position sets --position flag
func (ocb OutputCommandBuilder) Position(position Position) OutputCommandBuilder {
	ocb.monitor.Position = position

	return ocb
}

// LeftOf sets --left-of flag
func (ocb OutputCommandBuilder) LeftOf(monitor Monitor) OutputCommandBuilder {
	ocb.leftOf = &monitor

	return ocb
}

// Resolution sets --mode flag
func (ocb OutputCommandBuilder) Resolution(resolution Size) OutputCommandBuilder {
	ocb.monitor.Resolution = resolution

	return ocb
}

// MakePrimary sets --primary flag
func (ocb OutputCommandBuilder) MakePrimary() OutputCommandBuilder {
	ocb.monitor.Primary = true

	return ocb
}

// SetPrimary sets --primary flag
func (ocb OutputCommandBuilder) SetPrimary(primary bool) OutputCommandBuilder {
	ocb.monitor.Primary = primary

	return ocb
}

// EndOutput ends the output command builder inside command builder
func (ocb OutputCommandBuilder) EndOutput() CommandBuilder {
	ocb.parent.outputCommandBuilders = append(ocb.parent.outputCommandBuilders, ocb)

	return ocb.parent
}

func (ocb OutputCommandBuilder) getCommandArgs() ([]string, error) {
	var args []string

	if ocb.monitor.ID == "" {
		return args, nil
	}

	args = append(args, "--output", ocb.monitor.ID)
	if ocb.scale != 1 {
		args = append(args, "--scale", fmt.Sprintf("%0.3fx%0.3f", ocb.scale, ocb.scale))
	}

	args = append(args, "--pos", fmt.Sprintf("%dx%d", ocb.monitor.Position.X, ocb.monitor.Position.Y))

	mode, ok := ocb.monitor.CurrentMode()
	if !ok {
		return nil, fmt.Errorf(`cannot determin current mode for output "%s"`, ocb.monitor.ID)
	}

	res := mode.Resolution
	if res.Width > 0 && res.Height > 0 {
		args = append(args, "--mode", fmt.Sprintf("%dx%d", int(res.Width), int(res.Height)))
	}

	if ocb.leftOf != nil {
		args = append(args, "--left-of", ocb.leftOf.ID)
	}

	if ocb.monitor.Primary {
		args = append(args, "--primary")
	}

	return args, nil
}

func (cb CommandBuilder) getCommandArgs() []string {
	var args []string

	if cb.dpi > 0 {
		args = append(args, "--dpi", strconv.Itoa(cb.dpi))
	}

	if cb.screenSize.Width > 0 && cb.screenSize.Height > 0 {
		args = append(args, "--fb", fmt.Sprintf("%dx%d", int(cb.screenSize.Width), int(cb.screenSize.Height)))
	}

	return args
}

// RunnableCommands returns all the xrandr commands
func (cb CommandBuilder) RunnableCommands() ([]*exec.Cmd, error) {
	cmds := make([]*exec.Cmd, 0)
	args := cb.getCommandArgs()
	if len(args) > 0 {
		cmds = append(cmds, exec.Command("xrandr", cb.getCommandArgs()...))
	}

	for i, ocb := range cb.outputCommandBuilders {
		args, err := ocb.getCommandArgs()
		if err != nil {
			return nil, err
		}

		if i == 0 && len(cmds) > 0 {
			cmds[0].Args = append(cmds[0].Args, args...)
			continue
		}

		cmds = append(cmds, exec.Command("xrandr", args...))
	}

	return cmds, nil
}

// Run runs all the xrandr command returned by RunnableCommands method
func (cb CommandBuilder) Run() error {
	cmds, err := cb.RunnableCommands()
	if err != nil {
		return nil
	}

	for _, cmd := range cmds {
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// Command CommandBuilder constructor
func Command() CommandBuilder {
	return CommandBuilder{}
}
