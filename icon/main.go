package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"runtime"
	"strings"
	"utils"

	"fyne.io/systray"
)

var (
	broswerCmd *exec.Cmd
	desktopCmd *exec.Cmd
	logger     *slog.Logger
)

func main() {
	var (
		logFile *os.File
		err     error
	)

	mode := flag.Bool("dev", false, "specify if to build in production or development mode")
	flag.Parse()

	if logger, logFile, err = utils.Logger("trayIcon.log", *mode); err != nil {
		log.Fatalln(err)
	}
	defer logFile.Close()

	onExit := func() { broswerCmd, desktopCmd, logger = nil, nil, nil }
	systray.Run(onReady, onExit)
}

func onReady() {

	logger.Info("TrayIcon is alive!!!")

	systray.SetTemplateIcon(icon, icon)
	systray.SetTitle(title)
	systray.SetTooltip(toolTip)

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon, icon)
		systray.SetTitle(title)
		systray.SetTooltip(toolTip)

		launchBrowser := systray.AddMenuItem("Launch browser view", "Launch browser view")
		launchBrowserSubOne := launchBrowser.AddSubMenuItem("View in browser", "")
		launchBrowserSubTwo := launchBrowser.AddSubMenuItem("Close browser view", "")
		launchBrowserSubTwo.Disable()

		launchDesktop := systray.AddMenuItem("Launch desktop view", "Launch desktop view")

		systray.AddSeparator()
		logFolder := systray.AddMenuItem("Open log folder", "Open log folder")
		about := systray.AddMenuItem("More Information", "More Information")
		killDaemon := about.AddSubMenuItem("Quit daemon service", fmt.Sprintf("Quit %s daemon service", title))
		remove := about.AddSubMenuItem("Remove this tray-icon", "Remove this Icon")

		binaries := [3]string{"smDaemon", "smDesktop", "smBrowser"}
		for _, binary := range binaries {
			path := filepath.Join(getGOPATH(), "bin", binary)
			if _, err := os.Stat(path); err != nil {

				switch binary {
				case binaries[0]:
					utils.NotifyWithoutBeep(fmt.Sprintf("%s binary not found in GOPATH", title),
						fmt.Sprintf("%s binary not found in GOPATH, program cannot be initiated. Please see installation at %s", title, "https.github.com/tobigiwa/LiScreMon"))
					logger.Error(fmt.Sprintf("%s binary not found in GOPATH", title))
					systray.Quit()

				case binaries[1]:
					launchDesktop.Hide()
				case binaries[2]:
					launchBrowser.Hide()
				}
			}
		}

		for {
			select {
			case <-launchBrowserSubOne.ClickedCh:
				if strings.Contains(launchBrowser.String(), "running 🟢") {
					if err := jumpToBrowserView(); err != nil { // opens **another** browser tab of the `browser server's` port Addr
						utils.NotifyWithBeep("Operation failed", fmt.Sprintf("Could not launch %s browser view.", title))
						logger.Error(err.Error())
					}
					continue
				}

				if err := launchBrowserView(); err != nil { // starts the browser server
					utils.NotifyWithBeep("Operation failed", fmt.Sprintf("Could not launch %s browser view.", title))
					logger.Error(err.Error())
					continue
				}

				launchBrowserSubTwo.Enable()
				launchBrowser.SetTitle("Browser view: running 🟢")

			case <-launchBrowserSubTwo.ClickedCh: // closes the browser server
				broswerCmd.Process.Signal(os.Interrupt)
				broswerCmd.Wait()
				launchBrowserSubTwo.Disable()
				launchBrowser.SetTitle("Launch browser view")

			case <-launchDesktop.ClickedCh:
				if err := launcDesktopView(); err != nil { // desktop app is launched
					utils.NotifyWithBeep("Operation failed", fmt.Sprintf("Could not launch %s desktop view.", title))
					logger.Error(err.Error())
					continue
				}

				launchDesktop.Disable()
				launchDesktop.SetTitle("Desktop view: running 🟢")

				go func() {
					if err := desktopCmd.Wait(); err != nil { // waits for desktop app to be closed
						logger.Error(err.Error())
					}
					launchDesktop.Enable()
					launchDesktop.SetTitle("Launch desktop view")
				}()

			case <-logFolder.ClickedCh:
				if err := openLogFolder(); err != nil {
					utils.NotifyWithBeep("Operation failed", fmt.Sprintf("could not open log folder. Log folder is present at %s", utils.APP_LOGS_DIR))
					logger.Error(err.Error())
				}

			case <-killDaemon.ClickedCh:
				if err := killDaemonService(); err != nil { // starts the browser server
					utils.NotifyWithBeep("Operation failed", "Could not kill daemon service.")
					logger.Error(err.Error())
					continue
				}

				utils.NotifyWithBeep("Daemon sevice ended", "would need to restart user session to awake daemon.")

			case <-remove.ClickedCh:
				utils.NotifyWithBeep(
					"Uhmmm...Really? Buh why??? 😢",
					"You would have to restart your user session to have the 'Me' back. \nYou can still access the browser view (via the terminal) and the desktop app (via the desktop entry).\nI was just always a convenient option. Bye!!!👋")
				systray.Quit()
			}

		}
	}()
}

func launchBrowserView() error {

	gopath := getGOPATH()

	if runtime.GOOS == "linux" {
		gopathBin := filepath.Join(gopath, "bin", "smBrowser")
		broswerCmd = exec.Command(gopathBin)
	}

	if runtime.GOOS == "windows" {
		notImplemented()
	}

	return broswerCmd.Start()
}

func launcDesktopView() error {
	gopath := getGOPATH()

	if runtime.GOOS == "linux" {
		gopathBin := filepath.Join(gopath, "bin", "smDesktop")
		desktopCmd = exec.Command(gopathBin)
	}

	if runtime.GOOS == "windows" {
		notImplemented()
	}

	return desktopCmd.Start()
}

func notImplemented() {}

func jumpToBrowserView() error {
	var (
		portAddres string
		err        error
		cmd        *exec.Cmd
	)

	if portAddres, err = getBrowserRunningAddr(); err != nil {
		return err
	}

	if runtime.GOOS == "linux" {
		cmd = exec.Command("xdg-open", portAddres)
	}

	if runtime.GOOS == "windows" {
		notImplemented()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func getBrowserRunningAddr() (string, error) {
	byteData, err := os.ReadFile(utils.APP_JSON_CONFIG_FILE_PATH)
	if err != nil {
		return "", err
	}
	config, err := utils.DecodeJSON[utils.ConfigFile](byteData)
	if err != nil {
		return "", err
	}
	return config.BrowserAddr, nil
}

func getGOPATH() string {
	var gopath string
	if gopath = build.Default.GOPATH; gopath == "" {
		if gopath = os.Getenv("GOPATH"); gopath == "" {
			log.Fatalln("cannot build program, unable to determine GOPATH")
		}
	}
	return gopath
}

func killDaemonService() error {
	gopath := getGOPATH()
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" {
		gopathBin := filepath.Join(gopath, "bin", "smDaemon")
		cmd = exec.Command(gopathBin, "stop")
	}

	if runtime.GOOS == "windows" {
		notImplemented()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func openLogFolder() error {
	logFolder := utils.APP_LOGS_DIR

	var cmd *exec.Cmd

	if runtime.GOOS == "linux" {
		cmd = exec.Command("xdg-open", logFolder)
	}

	if runtime.GOOS == "windows" {
		notImplemented()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
