package monitoring

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
)

// currentlyOpenedWindows returns a list of all top-level windows.
func currentlyOpenedWindows(X *xgbutil.XUtil) ([]xproto.Window, error) {
	return ewmh.ClientListGet(X)
}

func setRootEventMask(X *xgbutil.XUtil) {
	err := xproto.ChangeWindowAttributesChecked(X.Conn(), X.RootWin(), xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskPropertyChange |
				xproto.EventMaskSubstructureNotify}).Check()
	if err != nil {
		log.Fatal("Failed to select notify events for root:", err)
	}
}

func registerWindowForEvents(windowID xproto.Window) {
	err := xproto.ChangeWindowAttributesChecked(x11Conn.Conn(), windowID, xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskStructureNotify}).Check()
	if err != nil {
		log.Fatalf("Failed to select notify events for window:%v, error: %v", windowID, err)
	}

	registerWindow(windowID)
}

// addWindowTocurSessionNamedWindowMap adds to the
/* curSessionNamedWindow map */
// there is no need to set Event mask on newly added windows and widows are not deleted from this map.
// This map is used to resolve the name of the known windows.
func addWindowTocurSessionNamedWindowMap(windowID xproto.Window, name string) {
	if _, exists := curSessionNamedWindow[windowID]; !exists {
		curSessionNamedWindow[windowID] = name
		return
	}
}

func getWindowClassName(X *xgbutil.XUtil, win xproto.Window) (string, error) {

	wmClass, err1 := xprop.PropValStrs(xprop.GetProperty(X, win, "WM_CLASS"))
	if err1 == nil && (len(wmClass) == 2) {
		return wmClass[1], nil
	}

	name, err2 := checkQueryTreeForParent(X, win)
	if err2 == nil && (name != "") {
		return name, nil
	}

	return "", fmt.Errorf("error on resolving name for window %d: %v, %w", win, err1, err2)
}

func checkQueryTreeForParent(X *xgbutil.XUtil, window xproto.Window) (string, error) {

	var (
		tree *xproto.QueryTreeReply
		err  error
	)

	if tree, err = xproto.QueryTree(X.Conn(), window).Reply(); err == nil {
		if parentName, ok := curSessionNamedWindow[tree.Parent]; ok && (tree.Parent != tree.Root) { // root would'nt have been added to curSessionNamedWindow
			return parentName, nil
		}

		ChildrenLen := len(tree.Children)
		for i := 0; i < ChildrenLen; i++ {
			if childName, ok := curSessionNamedWindow[tree.Children[i]]; ok { // noticed this behavior from vscode
				return childName, nil
			}
		}
	}
	return "", err
}

// needeAtom returns atom in the following other
//
// index 0: _NET_ACTIVE_WINDOW
//
// index 1: _NET_CLIENT_LIST_STACKING
func neededAtom() []xproto.Atom {
	netActiveWindowAtom, err := xprop.Atm(x11Conn, "_NET_ACTIVE_WINDOW")
	if err != nil {
		log.Fatalf("Could not get _NET_ACTIVE_WINDOW atom: %v", err)
	}
	netClientStackingAtom, err := xprop.Atm(x11Conn, "_NET_CLIENT_LIST_STACKING")
	if err != nil {
		log.Fatalf("Could not get _NET_CLIENT_LIST_STACKING atom: %v", err)
	}

	return []xproto.Atom{netActiveWindowAtom, netClientStackingAtom}
}
