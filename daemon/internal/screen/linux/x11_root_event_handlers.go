package monitoring

import (
	"fmt"
	"log"

	"time"
	"utils"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

func rootMapNotifyHandler(x11Conn *xgbutil.XUtil, ev xevent.MapNotifyEvent) {

	name, _ := getWindowClassName(x11Conn, ev.Window)
	registerWindowForEvents(ev.Window) // For DestroyNotify on the window

	if name != "" {
		addWindowTocurSessionNamedWindowMap(ev.Window, name)

	}
}

func (x11 *X11Monitor) windowChanged(x11Conn *xgbutil.XUtil, currActiveWindow xproto.Window) {

	formerActiveWindow := netActiveWindow

	x11.windowChangeCh <- utils.GenericKeyValue[xproto.Window, float64]{Key: formerActiveWindow.WindowID, Value: time.Since(formerActiveWindow.TimeStamp).Hours()}

	if formerActiveWindow.WindowName == "" { // this might be a not so needed check, cos -->
		formerActiveWindow.WindowName, _ = getWindowClassName(x11Conn, currActiveWindow) // NET_ACTIVE_WINDOW SHOULD ALWAYS HAVE A NAME, if not, that is lost metric
	}

	s := utils.ScreenTime{
		WindowID: formerActiveWindow.WindowID,
		AppName:  formerActiveWindow.WindowName,
		Type:     utils.Active,
		Duration: time.Since(formerActiveWindow.TimeStamp).Hours(),
		Interval: utils.TimeInterval{Start: formerActiveWindow.TimeStamp, End: time.Now()},
	}

	fmt.Printf("New active window ID =====> %v:%v\ntime elapsed for last window %v:%v was %vsecs\n",
		currActiveWindow, curSessionNamedWindow[currActiveWindow], formerActiveWindow.WindowID, curSessionNamedWindow[formerActiveWindow.WindowID], time.Since(netActiveWindow.TimeStamp).Seconds())

	// SETTING THE NEW _NET_ACTIVE_WINDOW
	netActiveWindow.WindowID = currActiveWindow                                   // SET THE WINDOW ID
	netActiveWindow.TimeStamp = time.Now()                                        // SET THE TIME
	netActiveWindow.WindowName, _ = getWindowClassName(x11Conn, currActiveWindow) // SET THE NAME

	if s.AppName != "" { // AS mentioned earlier, if we don't have a name, lost metric
		if err := x11.Db.WriteUsage(s); err != nil {
			log.Fatalln("write to db error:", err) // exit
		}
	}
}
