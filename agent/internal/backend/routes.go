package backend

import (
	views "agent/internal/frontend"
	"io/fs"
	"net/http"
	"runtime"
)

func (a *App) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	var staticFiles = fs.FS(views.AssetDir)
	staticFs, _ := fs.Sub(staticFiles, "assets")

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(staticFs))))

	// screentimePage
	mux.HandleFunc("GET /screentime", a.IndexPageHandler)
	mux.HandleFunc("GET /weekStat", a.WeekStatHandler)
	mux.HandleFunc("GET /appStat", a.AppStatHandler)
	mux.HandleFunc("GET /dayStat", a.DayStatHandler)
	mux.HandleFunc("GET /getCategory", a.getCategory)
	mux.HandleFunc("GET /back", a.backHandler)
	mux.HandleFunc("POST /setCategory", a.SetCategory)

	// ReminderAndLimitPage
	mux.HandleFunc("GET /tasks", a.tasksPage)
	mux.HandleFunc("GET /reminders", a.ReminderTasksHandler)
	mux.HandleFunc("GET /limits", a.limitTasksHandler)
	mux.HandleFunc("POST /newReminder", a.newReminderHandler)
	mux.HandleFunc("POST /newLimit", a.newDaillyAppLimitHandler)
	mux.HandleFunc("POST /removeTask/{uuid}", a.removeTask)

	switch runtime.GOOS {
	case "linux":
		mux.Handle("/tmp/", http.StripPrefix("/tmp/", http.FileServer(http.Dir("/tmp/"))))
	case "windows":
	case "darwin":
	default:
	}

	return mux
}
