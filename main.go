package main

import (
	"fmt"
	"os/exec"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
)

type Workspace struct {
	windows []xproto.Window
}

var (
	workspaces       []Workspace
	currentWorkspace int

	xu *xgbutil.XUtil
)

// EWMH
func initEWMH(xu *xgbutil.XUtil) error {
	// Set supported EWMH properties
	supported := []string{
		"_NET_WM_DESKTOP",
		"_NET_CURRENT_DESKTOP",
		"_NET_NUMBER_OF_DESKTOPS",
	}

	return ewmh.SupportedSet(xu, supported)
}

func setWindowWorkspace(window xproto.Window, workspace int, xu *xgbutil.XUtil) {
	ewmh.WmDesktopSet(xu, window, uint(workspace))
}

// Workspace func
func initWorkspace(n int) {
	workspaces = make([]Workspace, n)
	currentWorkspace = 0
}

func switchWorkspace(newWorkspace int, conn *xgb.Conn) {
	if newWorkspace < 0 || newWorkspace >= len(workspaces) {
		return
	}

	// Unmap windows in current workspace
	for _, window := range workspaces[currentWorkspace].windows {
		xproto.UnmapWindow(conn, window)
	}

	// Map windows in new workspace
	for _, window := range workspaces[newWorkspace].windows {
		xproto.MapWindow(conn, window)
		setWindowWorkspace(window, newWorkspace, xu)
	}

	currentWorkspace = newWorkspace
	ewmh.CurrentDesktopSet(xu, uint(newWorkspace))
}

func handleConfigureRequest(ev xproto.ConfigureRequestEvent, conn *xgb.Conn) {
	configureEvent := xproto.ConfigureNotifyEvent{
		Event:            ev.Window,
		Window:           ev.Window,
		AboveSibling:     0,
		X:                ev.X,
		Y:                ev.Y,
		Width:            ev.Width,
		Height:           ev.Height,
		BorderWidth:      ev.BorderWidth,
		OverrideRedirect: false,
	}

	xproto.SendEventChecked(
		conn, false, ev.Window, xproto.EventMaskStructureNotify, string(configureEvent.Bytes()))
}

// Map window and configure it
func handleMapRequest(ev xproto.MapRequestEvent, conn *xgb.Conn, screenWidth uint16, screenHeight uint16) (err error) {
	// Add the window to current workspace
	workspaces[currentWorkspace].windows = append(workspaces[currentWorkspace].windows, ev.Window)
	fmt.Println("Add window to current workspace")

	// Map window
	err = xproto.MapWindowChecked(conn, ev.Window).Check()
	if err != nil {
		return err
	}

	// Resize window to fullscreen
	values := []uint32{0, 0, uint32(screenWidth), uint32(screenHeight)}
	var masks uint16 = xproto.ConfigWindowX | xproto.ConfigWindowY | xproto.ConfigWindowWidth | xproto.ConfigWindowHeight
	xproto.ConfigureWindowChecked(conn, ev.Window, masks, values)

	// Focus window
	err = xproto.SetInputFocusChecked(conn, xproto.InputFocusPointerRoot, ev.Window, xproto.TimeCurrentTime).Check()
	if err != nil {
		return err
	}

	return nil
}

func handleKeyPress(ev xproto.KeyPressEvent, conn *xgb.Conn) {
	fmt.Println(ev.Detail)
	switch ev.Detail {
	case 65: // space
		exec.Command("rofi", "-show", "drun").Run()
	case 24: // q
		switchWorkspace(0, conn)
	case 25: // w
		switchWorkspace(1, conn)
	case 26: // e
		switchWorkspace(2, conn)
	case 27: // r
		switchWorkspace(3, conn)
	case 28: // t
		switchWorkspace(4, conn)
	case 29: // y
		switchWorkspace(5, conn)
	case 30: // u
		switchWorkspace(6, conn)
	case 31: // i
		switchWorkspace(7, conn)
	}
}

func main() {
	// Open connection to X
	conn, err := xgb.NewConn()
	if err != nil {
		fmt.Println("Could not open connection")
		return
	}

	// xgbutil
	xu, err = xgbutil.NewConn()
	if err != nil {
		fmt.Println("Could not open connnection to xgbutil")
		return
	}

	// Get conn info
	connInfo := xproto.Setup(conn)
	if connInfo == nil {
		fmt.Println("Could not parse connection info")
		return
	}

	screen := connInfo.DefaultScreen(conn)
	root := screen.Root

	initWorkspace(9)

	// Initialize ewmh
	err = initEWMH(xu)
	if err != nil {
		fmt.Println("Could not initialize EWMH")
		return
	}

	// Set attributes for root window
	mask := []uint32{
		xproto.EventMaskKeyPress |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskSubstructureRedirect,
	}
	err = xproto.ChangeWindowAttributesChecked(conn, root, xproto.CwEventMask, mask).Check()
	// Checks if another wm is running
	if err != nil {
		if _, ok := err.(xproto.AccessError); ok {
			fmt.Println("Another window manager is running")
			return
		}
	}

	// Event loop
	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			continue
		}
		if ev == nil && err == nil {
			break
		}

		switch event := ev.(type) {
		case xproto.KeyPressEvent:
			handleKeyPress(event, conn)
		case xproto.ConfigureRequestEvent:
			handleConfigureRequest(event, conn)
		case xproto.MapRequestEvent:
			handleMapRequest(event, conn, screen.WidthInPixels, screen.HeightInPixels)
		}
	}
}
