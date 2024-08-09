package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"strings"

	views "agent/internal/frontend/components"

	"github.com/google/uuid"

	utils "utils"
)

func (a *App) tasksPage(w http.ResponseWriter, r *http.Request) {

	var err error
	msg := utils.Message{
		Endpoint: strings.TrimPrefix(r.URL.Path, "/"),
	}

	msg, err = a.commWithDaemonService(msg)
	if err != nil {
		a.serverError(w, err)
		return
	}

	views.TasksPage(msg.TaskResponse.AllApps).Render(context.TODO(), w)
}

func (a *App) removeTask(w http.ResponseWriter, r *http.Request) {

	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		a.clientError(w, http.StatusBadRequest, errors.New(http.StatusText(http.StatusBadRequest)))
		return
	}

	endpoint, pathParam := path[1], r.PathValue("uuid")
	msg := utils.Message{
		Endpoint: endpoint,
	}

	var err error
	if msg.TaskRequest.UUID, err = uuid.Parse(pathParam); err != nil {
		a.clientError(w, http.StatusBadRequest, fmt.Errorf("bad task uuid format:%w", err))
		return
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		a.serverError(w, err)
		return
	}

	if r.URL.Query().Get("taskType") == "AppLimit" {
		http.Redirect(w, r, "/limits", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/reminders", http.StatusSeeOther)
}
