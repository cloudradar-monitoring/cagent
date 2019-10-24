// +build windows

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/cloudradar-monitoring/cagent"
)

type UI struct {
	MainWindow  *walk.MainWindow
	DataBinder  *walk.DataBinder
	SuccessIcon *walk.Icon
	ErrorIcon   *walk.Icon
	StatusBar   *walk.StatusBarItem
	SaveButton  *walk.ToolButton

	cagent           *cagent.Cagent
	installationMode bool
}

type setupErrors struct {
	connectionError error
	configError     error
	serviceError    error
}

func (se *setupErrors) SetConnectionError(err error) {
	se.connectionError = err
}

func (se *setupErrors) SetConfigError(err error) {
	se.configError = err
}

func (se *setupErrors) SetServiceError(err error) {
	se.serviceError = err
}

func (se *setupErrors) Describe() string {
	buf := new(bytes.Buffer)
	if se.connectionError != nil {
		fmt.Fprintf(buf, "Hub connection failed: %v", se.connectionError)
		return buf.String()
	} else {
		fmt.Fprintln(buf, "Hub connection succeeded.")
	}
	if se.configError != nil {
		fmt.Fprintf(buf, "Failed to save settings: %v", se.configError)
		return buf.String()
	} else {
		fmt.Fprintln(buf, "Your settings are saved.")
	}
	if se.serviceError != nil {
		fmt.Fprintf(buf, "Failed to start Cagent service: %v", se.serviceError)
		return buf.String()
	} else {
		fmt.Fprint(buf, "Services restarted and you are all set up!")
	}
	return buf.String()
}

// CheckSaveAndReload trying to test the Hub address and credentials from the config.
// If testOnly is true do not show alert message about the status (used to test the existing config on start).
func (ui *UI) CheckSaveAndReload(testOnly bool) {
	saveButtonText := ui.SaveButton.Text()
	defer func() {
		ui.SaveButton.SetText(saveButtonText)
		ui.SaveButton.SetEnabled(true)
	}()

	ui.SaveButton.SetEnabled(false)
	ui.SaveButton.SetText("Testing...")

	ctx := context.Background()
	setupStatus := &setupErrors{}
	err := ui.cagent.CheckHubCredentials(ctx, "URL", "User", "Password")
	if err != nil {
		if !testOnly {
			setupStatus.SetConnectionError(err)
			ui.StatusBar.SetText("Status: failed to connect to the Hub")
			ui.StatusBar.SetIcon(ui.ErrorIcon)
			RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		}
		return
	} else if testOnly {
		// in case we running this inside msi installer, just exit
		if ui.installationMode {
			os.Exit(0)
		}
		// otherwise - provide a feedback for user and set the status
		ui.StatusBar.SetText("Status: successfully connected to the Hub")
		ui.StatusBar.SetIcon(ui.SuccessIcon)
		return
	}

	ui.SaveButton.SetText("Saving...")

	ui.cagent.Config.MinValuableConfig.IOMode = cagent.IOModeHTTP
	err = cagent.SaveConfigFile(&ui.cagent.Config.MinValuableConfig, ui.cagent.ConfigLocation)
	if err != nil {
		setupStatus.SetConfigError(errors.Wrap(err, "Failed to write config file"))
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		return
	}

	m, err := mgr.Connect()
	if err != nil {
		setupStatus.SetServiceError(errors.Wrap(err, "Failed to connect to Windows Service Manager"))
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService("cagent")
	if err != nil {
		setupStatus.SetServiceError(errors.Wrap(err, "Failed to find Cagent service"))
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		return
	}
	defer s.Close()

	ui.SaveButton.SetText("Stopping the service...")

	if err := stopService(ctx, s); err != nil {
		setupStatus.SetServiceError(errors.Wrap(err, "Failed to stop Cagent service"))
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		return
	}

	ui.SaveButton.SetText("Starting the service...")
	if err := startService(ctx, s); err != nil {
		setupStatus.SetServiceError(errors.Wrap(err, "Failed to start Cagent service"))
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", setupStatus.Describe(), nil)
		return
	}

	ui.StatusBar.SetText("Status: successfully connected to the Hub")
	ui.StatusBar.SetIcon(ui.SuccessIcon)
	if ui.installationMode {
		RunDialog(ui.MainWindow, ui.SuccessIcon, "Success", setupStatus.Describe(), func() {
			os.Exit(0)
		})
	}
	RunDialog(ui.MainWindow, ui.SuccessIcon, "Success", setupStatus.Describe(), nil)
}

// windowsShowSettingsUI draws a window and waits until it will be closed.
// When installationMode is true, close the window after successful test&save.
func windowsShowSettingsUI(ca *cagent.Cagent, installationMode bool) {
	ui := UI{
		cagent:           ca,
		installationMode: installationMode,
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	ui.SuccessIcon, err = walk.NewIconFromFile(filepath.Join(exPath, "resources", "success.ico"))
	if err != nil {
		log.Fatal(err)
	}
	ui.ErrorIcon, err = walk.NewIconFromFile(filepath.Join(exPath, "resources", "error.ico"))
	if err != nil {
		log.Fatal(err)
	}

	labelFont := Font{PointSize: 12, Family: "Segoe UI"}
	inputFont := Font{PointSize: 12, Family: "Segoe UI"}
	buttonFont := Font{PointSize: 12, Family: "Segoe UI"}

	err = MainWindow{
		AssignTo: &ui.MainWindow,
		Title:    "Cagent",
		MinSize:  Size{360, 300},
		MaxSize:  Size{660, 300},

		DataBinder: DataBinder{
			AssignTo:       &ui.DataBinder,
			Name:           "config",
			DataSource:     ca.Config,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Hub Connection Credentials",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "URL:",
						Font: labelFont,
					},
					LineEdit{
						Text: Bind("HubURL"),
						Font: inputFont,
					},

					Label{
						Text: "User:",
						Font: labelFont,
					},
					LineEdit{
						Text: Bind("HubUser"),
						Font: inputFont,
					},

					Label{
						Text: "Pass:",
						Font: labelFont,
					},
					LineEdit{
						Text: Bind("HubPassword"),
						Font: inputFont,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ToolButton{
						MinSize:            Size{380, 35},
						AlwaysConsumeSpace: true,
						AssignTo:           &ui.SaveButton,
						Text:               "Test and Save",
						Font:               buttonFont,
						OnClicked: func() {
							ui.DataBinder.Submit()
							ui.CheckSaveAndReload(false)
						},
					},
				},
			},
		},
		StatusBarItems: []StatusBarItem{{
			AssignTo:  &ui.StatusBar,
			Width:     40,
			OnClicked: func() {},
		}},
	}.Create()

	if err != nil {
		panic(err)
	}

	go func() {
		ui.CheckPermissions()
		ui.CheckSaveAndReload(true)
	}()

	// disable window resize
	win.SetWindowLong(ui.MainWindow.Handle(), win.GWL_STYLE, win.WS_CAPTION|win.WS_SYSMENU)
	win.ShowWindow(ui.MainWindow.Handle(), win.SW_SHOW)
	ui.MainWindow.Run()
}

func startService(ctx context.Context, s *mgr.Service) error {
	err := s.Start("is", "manual-started")
	if err != nil {
		err = errors.Wrap(err, "could not schedule a service to start")
		return err
	}

	return waitServiceState(ctx, s, svc.Running)
}

func stopService(ctx context.Context, s *mgr.Service) error {
	status, err := s.Control(svc.Stop)
	if err != nil {
		if strings.Contains(err.Error(), "has not been started") {
			return nil
		}
		err = errors.Wrap(err, "could not schedule a service to stop")
		return err
	}
	if status.State == svc.Stopped {
		return nil
	}
	return waitServiceState(ctx, s, svc.Stopped)
}

// waitServiceState checks the current state of a service and waits until it will match
// the expectedState, or a context deadline appearing first.
func waitServiceState(ctx context.Context, s *mgr.Service, expectedState svc.State) error {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				err := errors.Wrap(ctx.Err(), "timeout waiting for service to stop")
				return err
			}
			return nil
		default:
			currentStatus, err := s.Query()
			if err != nil {
				err := errors.Wrap(err, "could not retrieve service status")
				return err
			}
			if currentStatus.State == expectedState {
				return nil
			}
			time.Sleep(300 * time.Millisecond)
		}
	}
	return nil
}

func (ui *UI) CheckPermissions() {
	if checkIsRunningWithElevatedPrivileges() {
		return
	}
	RunDialog(ui.MainWindow, ui.ErrorIcon,
		"Error", "Please run this program with administrator priveleges.", func() {
			os.Exit(1)
		})
}

func checkIsRunningWithElevatedPrivileges() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

func RunDialog(owner walk.Form, icon *walk.Icon, title, text string, callback func()) (int, error) {
	var dlg *walk.Dialog
	var acceptPB *walk.PushButton
	font := Font{PointSize: 12, Family: "Segoe UI"}

	return Dialog{
		FixedSize:     true,
		AssignTo:      &dlg,
		Title:         title,
		DefaultButton: &acceptPB,
		MaxSize:       Size{320, 180},
		Font:          font,
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ImageView{
						Image: icon,
					},
					VSpacer{},
					TextLabel{
						MaxSize: Size{320, 180},
						Text:    text,
						Font:    font,
					},
				},
			},
			HSpacer{},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					PushButton{
						Font:     font,
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							dlg.Accept()
							if callback != nil {
								callback()
							}
						},
					},
				},
			},
		},
	}.Run(owner)
}
