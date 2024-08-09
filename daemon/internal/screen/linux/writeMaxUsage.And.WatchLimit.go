package monitoring

import (
	"context"
	"fmt"
	"log"

	"time"
	utils "utils"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/google/uuid"
)

func (x11 *X11Monitor) CloseWindowChangeCh() {
	close(x11.windowChangeCh)
}

func (x11 *X11Monitor) WindowChangeTimerFunc(ctx context.Context, timer *time.Timer) {
	for {
		select {
		case <-ctx.Done():
			return

		case t := <-x11.windowChangeCh:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(fixtyEight)
			x11.watchLimit(t.Key, t.Value)

		case <-timer.C:
			timer.Reset(fixtyEight)

			if netActiveWindow.WindowID == xevent.NoWindow {
				continue
			}

			x11.watchLimit(netActiveWindow.WindowID, fixtyEight.Hours())
			x11.sendFiftyEightSecsUsage()
		}
	}
}

func (x11 *X11Monitor) watchLimit(windowID xproto.Window, duration float64) {

	if windowName, ok := curSessionNamedWindow[windowID]; ok {
		if limitApp, ok := LimitApp[windowName]; ok {

			limitApp.timeSofar += duration

			if limitApp.timeSofar >= limitApp.limit { // limit reached
				delete(LimitApp, windowName)
				if err := x11.appLimitReached(limitApp.taskUUID); err != nil {
					x11.logger.Error(err.Error())
				}

				return
			}

			if timeLeft := limitApp.limit - limitApp.timeSofar; timeLeft > float64(0.125) && timeLeft <= float64(0.1666) && !limitApp.tenMinToLimit {
				limitApp.tenMinToLimit = true
				LimitApp[windowName] = limitApp
				if err := x11.appLimitNotification(limitApp.taskUUID, "10"); err != nil {
					x11.logger.Error(err.Error())
				}

			} else if timeLeft > float64(0.0583) && timeLeft <= float64(0.0833) && !limitApp.fiveMinToLimit {
				limitApp.fiveMinToLimit = true
				LimitApp[windowName] = limitApp
				if err := x11.appLimitNotification(limitApp.taskUUID, "5"); err != nil {
					x11.logger.Error(err.Error())
				}
			}

			LimitApp[windowName] = limitApp
		}
	}
}

func (x11 *X11Monitor) sendFiftyEightSecsUsage() {

	oneMinuteUsage := time.Since(netActiveWindow.TimeStamp).Hours()
	oneMinuteTimeStamp := netActiveWindow.TimeStamp

	netActiveWindow.TimeStamp = time.Now()

	if err := x11.Db.WriteUsage(utils.ScreenTime{
		WindowID: netActiveWindow.WindowID,
		AppName:  netActiveWindow.WindowName,
		Type:     utils.Active,
		Duration: oneMinuteUsage,
		Interval: utils.TimeInterval{Start: oneMinuteTimeStamp, End: time.Now()},
	}); err != nil {
		log.Fatalln("write to db error:", err) // exit
	}
}

var LimitApp = make(map[string]limitWindow, 20)

func AddNewLimit(t utils.Task, timesofar float64) {
	if timesofar < t.AppLimit.Limit {
		LimitApp[t.AppName] = limitWindow{
			windowInfo: windowInfo{WindowName: t.AppName},
			taskUUID:   t.UUID,
			limit:      t.AppLimit.Limit,
			timeSofar:  timesofar,
		}
	}
}

func (x11 *X11Monitor) appLimitReached(taskID uuid.UUID) error {

	task, err := x11.Db.GetTaskByUUID(taskID)
	if err != nil {
		return err
	}

	title := fmt.Sprintf("Usage Limit reached for %s", task.AppName)
	subtitle := fmt.Sprintf("App: %s Usage Limit: %s", task.AppName, utils.UsageTimeInHrsMin(task.AppLimit.Limit))

	utils.NotifyWithBeep(title, subtitle)

	if task.AppLimit.OneTime {
		return x11.Db.RemoveTask(taskID)
	}

	return nil
}

func (x11 *X11Monitor) appLimitNotification(taskID uuid.UUID, left string) error {
	task, err := x11.Db.GetTaskByUUID(taskID)
	if err != nil {
		return err
	}

	title := fmt.Sprintf("%s minute usage left for %s", left, task.AppName)
	subtitle := fmt.Sprintf("App: %s; Usage Limit: %s", task.AppName, utils.UsageTimeInHrsMin(task.AppLimit.Limit))

	utils.NotifyWithoutBeep(title, subtitle)

	return nil
}
