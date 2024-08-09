package monitoring

import (
	"fmt"
	"log/slog"
	"time"

	db "smDaemon/daemon/internal/database"

	"utils"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

var (
	// curSessionNamedWindow is a map of all current session "named" windows.
	// An X session is typically a time between login and logout (or restart/shutdown).
	// Only windows with knowm WM_CLASS are added to this map. The X_ID are always unique
	// for a particular window in each session.
	curSessionNamedWindow = make(map[xproto.Window]string, 20)
	netActiveWindowAtom   xproto.Atom
	monitor               X11Monitor

	// netClientStackingAtom xproto.Atom
	x11Conn         *xgbutil.XUtil
	netActiveWindow = &netActiveWindowInfo{}
	fixtyEight  = time.Duration(58) * time.Second
)

func InitMonitoring(db *db.BadgerDBStore, logger *slog.Logger) (X11Monitor, error) {

	var err error

	count := 0
	// X server connection
	for {
		if x11Conn, err = xgbutil.NewConn(); err != nil { // we wait till we connect to X server
			count++
			if count > 10 { // 10 retries
				return X11Monitor{}, fmt.Errorf("error connecting to X server:%w", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	utils.X11Connection = x11Conn

	monitor = X11Monitor{
		X11Connection:  x11Conn,
		Db:             db,
		windowChangeCh: make(chan utils.GenericKeyValue[xproto.Window, float64], 1),
		logger:         logger,
	}

	if err = setRootEventMask(x11Conn); err != nil {
		return X11Monitor{}, err
	}

	registerRootWindowForEvents(x11Conn)

	windows, err := currentlyOpenedWindows(x11Conn)
	if err != nil {
		return X11Monitor{}, err
	}

	for _, window := range windows {
		getWindowClassName(x11Conn, window)
		registerWindowForEvents(window)
	}
	atoms, err := neededAtom()
	if err != nil {
		return X11Monitor{}, err
	}

	netActiveWindowAtom, _ = atoms[0], atoms[1]
	netActiveWindow.WindowID = xevent.NoWindow

	return monitor, nil
}
