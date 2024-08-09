// `Package service` is the package that houses the
// functionalites that attends to the need of the agent.
// It can depend on all other packages/directories in `internal`.
package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	db "smDaemon/daemon/internal/database"
	"smDaemon/daemon/internal/tasks"

	"syscall"
	utils "utils"
)

func NewService(db *db.BadgerDBStore) (*Service, error) {

	var (
		service Service
		err     error
	)

	service.db = db

	if service.taskManager, err = tasks.StartTaskManger(db); err != nil {
		return nil, fmt.Errorf("error starting task manager")
	}

	return &service, nil
}

func (s *Service) StartService(socketDir string, db *db.BadgerDBStore) error {

	listener, err := unixDomainSocket(socketDir)
	if err != nil {
		return err
	}
	s.handleConnection(listener) // blocking

	listener.Close()
	return nil
}

func unixDomainSocket(socketDir string) (*net.UnixListener, error) {

	var (
		conn     *net.UnixListener
		err      error
		unixAddr *net.UnixAddr
	)

	if err = os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating socket dir:%w", err)
	}

	socketFilePath := filepath.Join(socketDir, "daemon.sock")

	syscall.Unlink(socketFilePath)

	if unixAddr, err = net.ResolveUnixAddr("unix", socketFilePath); err != nil {
		return nil, err
	}

	if conn, err = net.ListenUnix("unix", unixAddr); err != nil {
		return nil, fmt.Errorf("error creating domain socket:%w", err)
	}

	conn.SetUnlinkOnClose(true)

	return conn, nil
}

func (s *Service) handleConnection(listener *net.UnixListener) {
	for {
		c, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("connection closed, daemonService says 🖐")
				return
			} else {
				log.Println("error accepting connection...buh we muuve:", err)
				continue
			}
		}

		log.Println("Connection accepted")
		go s.treatMessage(c)
	}
}

func (s *Service) treatMessage(c net.Conn) {
	for {
		var (
			msg utils.Message
			err error
			n   int
		)

		buf := make([]byte, 10_000) //10kb
		if n, err = c.Read(buf); err != nil {
			log.Println("error reading message:", err)
			if errors.Is(err, io.EOF) {
				log.Println("client connection closed")
				c.Close()
				return
			}
			continue
		}

		if msg, err = utils.DecodeJSON[utils.Message](buf[:n]); err != nil {
			log.Println("error decoding socket message", err)
			c.Close()
			return
		}

		switch msg.Endpoint {

		case "startConnection":
			msg = utils.Message{StatusCheck: "OK"}

		case "closeConnection":
			closeConnection(c)
			return

		case "weekStat":
			msg.WeekStatResponse, err = s.getWeekStat(msg.WeekStatRequest, false)

		case "appStat":
			msg.AppStatResponse, err = s.getAppStat(msg.AppStatRequest)

		case "dayStat":
			msg.DayStatResponse, err = s.getDayStat(msg.DayStatRequest)

		case "setCategory":
			msg.SetCategoryResponse, err = s.setAppCategory(msg.SetCategoryRequest)

		case "getCategory":
			msg.GetCategoryResponse, err = s.getCategory()

		case "tasks":
			msg.TaskResponse, err = s.tasks()

		case "reminders":
			msg.TaskResponse, err = s.allReminderTask()

		case "limits":
			msg.TaskResponse, err = s.allDailyAppLimitTask()

		case "newReminder":
			msg.TaskResponse, err = s.addNewReminder(msg.TaskRequest)

		case "newLimit":
			msg.TaskResponse, err = s.addNewLimitApp(msg.TaskRequest)

		case "removeTask":
			msg.TaskResponse, err = s.removeTask(msg.TaskRequest)

		case "back":
			msg.WeekStatResponse, err = s.getWeekStat(msg.WeekStatRequest, true)

		}

		if err != nil {
			msg.IsError, msg.Error = true, err.Error()
		}

		bytes, err := utils.EncodeJSON(msg)
		if err != nil {
			msg.Error = fmt.Sprintf("error encoding response in serviceInstance: %s: %s", msg.Error, err.Error())
		}

		_, err = c.Write(bytes)
		if err != nil {
			log.Println("error writing response:", err)
			continue
		}
	}
}

func closeConnection(c net.Conn) {
	fmt.Println("we got a close connection message")
	err := c.Close()
	if err != nil {
		log.Println("ERROR CLOSING CONNECTION")
	}
}
