package daemon

import (
	db "LiScreMon/daemon/internal/database"
	monitoring "LiScreMon/daemon/internal/screen/linux"
	"LiScreMon/daemon/internal/service"
	"path/filepath"

	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	helperFuncs "pkg/helper"
	"pkg/types"

	"github.com/BurntSushi/xgbutil/xevent"
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("error at init fn:", err) // exit
	}

	configDir := filepath.Join(homeDir, "liScreMon")
	logDir := filepath.Join(configDir, "logs")

	for _, dir := range [2]string{configDir, logDir} {
		if err = os.MkdirAll(dir, 0755); err != nil {
			log.Fatalln("error at init fn:", err) // exit
		}
	}

	jsonConfigFile := filepath.Join(configDir, "config.json")
	file, err := os.Create(jsonConfigFile)
	if err != nil {
		log.Fatalln("error at init fn:", err) // exit
	}

	byteData, err := helperFuncs.EncodeJSON(types.ConfigFile{Name: "LiScreMon", Description: "Linux Screen Monitoring", Version: "1.0.0"})
	if err != nil {
		log.Fatalln("error at init fn:", err) // exit
	}

	file.Write(byteData)
	file.Close()
}

func DaemonServiceLinux(logger *slog.Logger) {

	// config directory
	configDir, err := helperFuncs.ConfigDir()
	if err != nil {
		log.Fatalln(err) // exit
	}
	// database
	badgerDB, err := db.NewBadgerDb(filepath.Join(configDir, "badgerDB"))
	if err != nil {
		log.Fatalln(err) // exit
	}

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// service
	service, err := service.NewService(badgerDB)
	if err != nil {
		log.Fatalln(err) // exit
	}

	go func() {
		if err := service.StartService(filepath.Join(configDir, "socket"), badgerDB); err != nil {
			log.Println("error starting service", err)
			sig <- syscall.SIGTERM // if service.StartService fails, send a signal to close the program
		}
	}()

	monitor, err := monitoring.InitMonitoring(badgerDB)
	if err != nil {
		log.Fatalln(err) // exit
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	timer := time.NewTimer(time.Duration(58) * time.Second)

	go func() {
		monitor.WindowChangeTimerFunc(ctx, timer)
	}()

	go func() {
		xevent.Main(monitor.X11Connection) // Start the x11 event loop.
		log.Println("error starting x11 event loop", err)
		sig <- syscall.SIGTERM // if the event loop cannot be started, send a signal to close the program
	}()

	<-sig // awaiting only the first signal

	// err = monitor.Db.UpdateAppInfoManually([]byte("app:Google-chrome"), db.ExampleOf_opsFunc)
	// if err != nil {
	// 	fmt.Println("opt failed", err)
	// }

	xevent.Quit(monitor.X11Connection) // this should always comes first
	ctxCancel()                        // a different goroutine for managing backing up app usage every minute, fired from monitor
	monitor.CloseWindowChangeCh()      // a different goroutine,closes a channel, this should be after calling the CancelFunc passed to monitor.WindowChangeTimerFunc

	if !timer.Stop() {
		<-timer.C
	}

	service.StopTaskManger() // a different goroutine for managing taskManager, fired from service
	badgerDB.Close()
	close(sig)
}
