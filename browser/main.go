package main

import (
	webserver "agent"
	"flag"
	"fmt"
	"net"

	"runtime"
	"time"
	"utils"

	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {

	mode := flag.Bool("dev", false, "specify if to build in production or development mode")
	flag.Parse()

	// logging
	logger, logFile, err := utils.Logger("webserver.log", *mode)
	if err != nil {
		log.Fatalln(err) // exit
	}
	defer logFile.Close()

	slog.SetDefault(logger)

	var count, port int
	for {
		count++
		if port, err = findFreePort(*mode); err != nil {
			if count >= 5 {
				log.Fatalf("error getting a free port for browser connection: err %v", err) // exit
			}
			time.Sleep(time.Second)
			continue
		}
		break
	}

	BrowserAgent, err := webserver.BrowserAgent(logger)
	if err != nil {
		log.Fatalln("error creating BrowserAgent:", err) // exit
	}

	_, err = BrowserAgent.CheckDaemonService()
	if err != nil {
		log.Fatalln("error connecting to daemon service:", err) // exit
	}

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", port),
		Handler:  BrowserAgent.Routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	url := fmt.Sprintf("http://127.0.0.1:%d/screentime", port)

	_ = writeURLtoJSONConfigFile(url)

	go func() {
		log.Println("Server is running on " + url)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error: " + err.Error())
		}
	}()

	time.Sleep(200 * time.Millisecond) // waiting for the server to be up and running
	cmd := openURLInBrowser(url)
	cmd.Start()

	<-done
	close(done)

	if err := cmd.Wait(); err != nil {
		logger.Error("err with browser launch command:" + err.Error())
	}

	if err := BrowserAgent.CloseDaemonConnection(); err != nil {
		logger.Error("error closing socket connection with daemon, error:" + err.Error())
	}

	if err := server.Shutdown(context.TODO()); err != nil {
		logger.Error("Graceful server shutdown Failed: " + err.Error())
	}

	logger.Info("SERVER STOPPED GRACEFULLY")
}

func findFreePort(devMode bool) (int, error) {
	if devMode {
		return 8080, nil
	}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	// Type assert the Addr to a *net.TCPAddr to extract the port number
	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}

// OpenURLInBrowser would return nil if OS is not linux/windows/darwin
func openURLInBrowser(url string) *exec.Cmd {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		return nil
	}

	args = append(args, url)
	return exec.Command(cmd, args...)
}

func writeURLtoJSONConfigFile(url string) error {
	configFile := utils.APP_JSON_CONFIG_FILE_PATH

	byteData, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	config, err := utils.DecodeJSON[utils.ConfigFile](byteData)
	if err != nil {
		return err
	}

	config.BrowserAddr = url

	if byteData, err = utils.EncodeJSON(config); err != nil {
		return err
	}

	return os.WriteFile(configFile, byteData, 0644)
}
