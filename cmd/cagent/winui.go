// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/cloudradar-monitoring/cagent"
)

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

type UI struct {
	MainWindow  *walk.MainWindow
	DataBinder  *walk.DataBinder
	SuccessIcon *walk.Icon
	ErrorIcon   *walk.Icon
	StatusBar   *walk.StatusBarItem
	SaveButton  *walk.ToolButton

	installationMode bool
	ca               *cagent.Cagent
}

// TestSaveReload trying to test the HUB address and credentials from the config
// if testOnly is true do not show alert message about the status(used to test the existed config on start)
func (ui *UI) TestSaveReload(testOnly bool) {
	// it will become messy if we will handle all UI errors here
	// just ignore them and go further, because next steps don't depend on previous ones

	saveButtonText := ui.SaveButton.Text()
	ui.SaveButton.SetEnabled(false)
	ui.SaveButton.SetText("Testing...")
	err := ui.ca.TestHubWinUI()
	defer func() {
		ui.SaveButton.SetText(saveButtonText)
		ui.SaveButton.SetEnabled(true)
	}()

	if err != nil {
		// do not set the error in the status bar by default
		if err == cagent.ErrorTestWinUISettingsAreEmpty && testOnly {
			return
		}

		ui.StatusBar.SetText("Status: failed to connect to the HUB ")
		ui.StatusBar.SetIcon(ui.ErrorIcon)

		if !testOnly {
			RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", err.Error(), nil)
		}
		return
	}

	if testOnly {
		// in case we running this inside msi installer, just exit
		if ui.installationMode {
			os.Exit(0)
		}

		// otherwise - provide a feedback for user and set the status
		ui.StatusBar.SetText("Status: successfully connected to the HUB")
		ui.StatusBar.SetIcon(ui.SuccessIcon)
		return
	}

	message := "Test connection succeeded.\n"

	ui.SaveButton.SetText("Saving...")
	err = cagent.SaveConfigFile(&ui.ca.Config.MinValuableConfig, ui.ca.ConfigLocation)
	if err != nil {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", message+"Failed to save config: "+err.Error(), nil)
		return
	}

	message += "Your settings are saved.\n"

	m, err := mgr.Connect()
	if err != nil {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", message+"Failed to connect to Windows Service Manager: "+err.Error(), nil)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService("cagent")
	if err != nil {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", message+"Failed to connect to find 'cagent' service: "+err.Error(), nil)
		return
	}
	defer s.Close()

	ui.SaveButton.SetText("Stopping the service...")
	err = stopService(s)
	if err != nil && !strings.Contains(err.Error(), "has not been started") {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", message+"Failed to stop 'cagent' service: "+err.Error(), nil)
		return
	}

	ui.SaveButton.SetText("Starting the service...")
	err = startService(s)
	if err != nil {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", message+"Failed to start 'cagent' service: "+err.Error(), nil)
		return
	}

	var callback func()

	if ui.installationMode {
		callback = func() {
			os.Exit(0)
		}
	}

	RunDialog(ui.MainWindow, ui.SuccessIcon, "Success", message+"Services restarted and you are all set up!", callback)
	ui.StatusBar.SetText("Status: successfully connected to the HUB")
	ui.StatusBar.SetIcon(ui.SuccessIcon)
}

// windowsShowSettingsUI draws a window and wait until it will be closed closed
// when installationMode is true close the window after successful test&save
func windowsShowSettingsUI(ca *cagent.Cagent, installationMode bool) {
	ui := UI{ca: ca, installationMode: installationMode}
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

	MainWindow{
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
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "HUB URL",
						Font: labelFont,
					},
					LineEdit{
						Text: Bind("HubURL"),
						Font: inputFont,
					},

					Label{
						Text: "HUB USER",
						Font: labelFont,
					},
					LineEdit{
						Text: Bind("HubUser"),
						Font: inputFont,
					},

					Label{
						Text: "HUB PASSWORD",
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
							ui.TestSaveReload(false)
						},
					},
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			{
				AssignTo: &ui.StatusBar,
				Width:    40,
				OnClicked: func() {
					// todo: show full error on click
				},
			},
		},
	}.Create()

	go func() {
		ui.ShowAdminAlert()
		ui.TestSaveReload(true)
	}()

	// disable window resize
	win.SetWindowLong(ui.MainWindow.Handle(), win.GWL_STYLE, win.WS_CAPTION|win.WS_SYSMENU)
	win.ShowWindow(ui.MainWindow.Handle(), win.SW_SHOW)
	ui.MainWindow.Run()
}

func startService(s *mgr.Service) error {
	err := s.Start("is", "manual-started")
	if err != nil {
		return err
	}

	return waitServiceState(s, svc.Status{}, svc.Running, time.Second*15)
}

func stopService(s *mgr.Service) error {
	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}

	return waitServiceState(s, status, svc.Stopped, time.Second*15)
}

func waitServiceState(s *mgr.Service, currentStatus svc.Status, state svc.State, timeout time.Duration) error {
	stopAt := time.Now().Add(timeout)
	var err error
	for currentStatus.State != state {
		if stopAt.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to stop")
		}
		time.Sleep(300 * time.Millisecond)
		currentStatus, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}

	return nil
}

func (ui *UI) ShowAdminAlert() {
	if !checkAdmin() {
		RunDialog(ui.MainWindow, ui.ErrorIcon, "Error", "Please run as administrator", func() { os.Exit(1) })
	}
}

func checkAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}
