package backend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	views "agent/internal/frontend/components"
	utils "utils"

	"github.com/google/uuid"
)

func (a *App) limitTasksHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	msg := utils.Message{
		Endpoint: strings.TrimPrefix(r.URL.Path, "/"),
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		a.serverError(w, err)
		return
	}

	if err = views.RenderTasks(views.AppLimitTasks(msg.TaskResponse.AllTask)).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}

func (a *App) newDaillyAppLimitHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		a.clientError(w, http.StatusBadRequest, err)
		return
	}

	var (
		msg      utils.Message
		task     utils.Task
		hrs, min int
		err      error
	)

	task.Job = utils.DailyAppLimit
	task.AppLimit.OneTime = true

	for key, value := range r.Form {
		switch key {
		case "app":
			task.AppName = value[0]

		case "hrs":
			if value[0] == "" {
				hrs = 0
				continue
			}

			if hrs, err = strconv.Atoi(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}

		case "min":
			if value[0] == "" {
				min = 0
				continue
			}

			if min, err = strconv.Atoi(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}

		case "recurring":
			if _, err := strconv.ParseBool(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.AppLimit.OneTime = false

		case "exitApp":
			if _, err := strconv.ParseBool(value[0]); err != nil {
				a.clientError(w, http.StatusBadRequest, fmt.Errorf("error parsing formData:%w", err))
				return
			}
			task.AppLimit.ExitApp = true

		}
	}

	hours, minutes := time.Duration(hrs)*time.Hour, time.Duration(min)*time.Minute

	if task.AppLimit.Limit = hours.Hours() + minutes.Hours(); task.AppLimit.Limit <= 0 {
		a.clientError(w, http.StatusBadRequest, fmt.Errorf("cannot watch for zero-time limit"))
		return
	}

	task.AppLimit.Day = utils.Today()
	task.UUID = uuid.New()

	msg = utils.Message{
		Endpoint:    strings.TrimPrefix(r.URL.Path, "/"),
		TaskRequest: task,
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		if strings.Contains(err.Error(), utils.ErrLimitAppExist.Error()) {
		}
		a.serverError(w, err)
		return
	}

	if err = views.RenderTasks(views.AppLimitTasks(msg.TaskResponse.AllTask)).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}
