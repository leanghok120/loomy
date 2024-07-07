package main

import (
	"fmt"
	"os/exec"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

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
	// Map window
	err = xproto.MapWindowChecked(conn, ev.Window).Check()
	if err != nil {
		return err
	}

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

func handleKeyPress(ev xproto.KeyPressEvent) {
	fmt.Println(ev.Detail)
	// 'p'
	if ev.Detail == 33 {
		exec.Command("rofi", "-show", "drun").Run()
	}
}

func main() {
	// Open connection to X
	conn, err := xgb.NewConn()
	if err != nil {
		fmt.Println("Could not open connection")
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
			handleKeyPress(event)
		case xproto.ConfigureRequestEvent:
			handleConfigureRequest(event, conn)
		case xproto.MapRequestEvent:
			handleMapRequest(event, conn, screen.WidthInPixels, screen.HeightInPixels)
		}
	}
}
