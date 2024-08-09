package backend

import (
	views "agent/internal/frontend/components"
	"context"
	"net/http"
)

func (a *App) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := views.IndexPage().Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
		return
	}
}
