package backend

import (
	views "agent/internal/frontend/components"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	utils "utils"

	"github.com/google/uuid"
)

func (a *App) ReminderTasksHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	msg := utils.Message{
		Endpoint: strings.TrimPrefix(r.URL.Path, "/"),
	}

	msg, err = a.commWithDaemonService(msg)
	if err != nil {
		a.serverError(w, err)
		return
	}

	if err = views.RenderTasks(views.ReminderTasks(msg.TaskResponse.AllTask)).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}

func (a *App) newReminderHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		a.clientError(w, http.StatusBadRequest, err)
		return
	}

	var (
		task utils.Task
		err  error
	)

	for key, value := range r.Form {
		switch key {
		case "title":
			task.UI.Title = value[0]

		case "subtitle":
			task.UI.Subtitle = value[0]

		case "reminder":
			startTime, err := time.ParseInLocation("2006-01-02T15:04", value[0], time.Local)
			if err != nil {
				a.clientError(w, http.StatusBadRequest, err)
				return
			}
			task.Reminder.StartTime = startTime

		case "soundNotification":
			if _, err := strconv.ParseBool(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.Reminder.AlertSound[2] = true

		case "firstNotification":
			alert, err := strconv.Atoi(value[0])
			if err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.Reminder.AlertTimesInMinutes[0] = alert

		case "soundFirstNotification":
			if _, err := strconv.ParseBool(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.Reminder.AlertSound[0] = true

		case "secondNotification":
			alert, err := strconv.Atoi(value[0])
			if err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.Reminder.AlertTimesInMinutes[1] = alert

		case "soundSecondNotification":
			if _, err := strconv.ParseBool(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.Reminder.AlertSound[1] = true

		case "app":
			if task.AppName = value[0]; value[0] == "no-app" {
				task.Job = utils.ReminderWithNoAppLaunch
			} else {
				task.Job = utils.ReminderWithAppLaunch
			}

		case "note":
			task.UI.Notes = value[0]
		}
	}

	task.UUID = uuid.New()
	msg := utils.Message{
		Endpoint:    strings.TrimPrefix(r.URL.Path, "/"),
		TaskRequest: task,
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		a.serverError(w, err)
		return
	}

	if err = views.RenderTasks(views.ReminderTasks(msg.TaskResponse.AllTask)).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}
