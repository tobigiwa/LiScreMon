package daemon

import (
	"path/filepath"
	db "smDaemon/daemon/internal/database"
	monitoring "smDaemon/daemon/internal/screen/linux"
	"smDaemon/daemon/internal/service"

	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"utils"

	"github.com/BurntSushi/xgbutil/xevent"
)

func DaemonServiceLinux(logger *slog.Logger) error {

	// config directory
	configDir := utils.APP_CONFIG_DIR

	// database
	badgerDB, err := db.NewBadgerDb(filepath.Join(configDir, "badgerDB"))
	if err != nil {
		return err
	}

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// service
	service, err := service.NewService(badgerDB)
	if err != nil {
		return err
	}

	go func() {
		if err := service.StartService(filepath.Join(configDir, "socket"), badgerDB); err != nil {
			logger.Error("error starting service:" + err.Error())
			sig <- syscall.SIGTERM // if service.StartService fails, send a signal to close the program
		}
	}()

	monitor, err := monitoring.InitMonitoring(badgerDB, logger)
	if err != nil {
		return err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	timer := time.NewTimer(time.Duration(58) * time.Second)

	go func() {
		monitor.WindowChangeTimerFunc(ctx, timer)
	}()

	go func() {
		// Start the x11 event loop.
		xevent.Main(monitor.X11Connection)
		logger.Error("error starting x11 event loop:" + err.Error())
		sig <- syscall.SIGTERM // if the event loop cannot be started, send a signal to close the program
	}()

	<-sig // awaiting only the first signal

	// err = monitor.Db.UpdateOpertionOnPrefix("app", db.ExampleOf_opsFunc)
	// if err != nil {
	// 	log.Println("opt failed", err)
	// }
	// err = monitor.Db.UpdateOperationOnKey([]byte("Microsoft-edge"), db.ExampleOf_opsFunc)
	// if err != nil {
	// 	log.Println("opt failed", err)
	// }

	xevent.Quit(monitor.X11Connection) // this should always comes first
	ctxCancel()                        // closes the goroutine fired for `monitor.WindowChangeTimerFunc(ctx, timer)`
	monitor.CloseWindowChangeCh()      // closes the channel used in `monitor.WindowChangeTimerFunc(ctx, timer)`, this should be after calling the CancelFunc passed to `monitor.WindowChangeTimerFunc`

	if !timer.Stop() {
		<-timer.C
	}

	if err := service.StopTaskManger(); err != nil { // a different goroutine for managing taskManager, fired from service
		logger.Error(err.Error())
	}

	if err := badgerDB.Close(); err != nil {
		logger.Error(err.Error())
	}

	close(sig)
	return nil
}
