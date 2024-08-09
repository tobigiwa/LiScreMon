package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"utils"

	views "agent/internal/frontend/components"
	"strings"
)

var lastAppInfos = make([]utils.AppIconCategoryAndCmdLine, 0, 30)

func (a *App) SetCategory(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	category := r.Form.Get("category")
	appName := r.URL.Query().Get("appName")

	fmt.Println(appName, category)

	msg := utils.Message{
		Endpoint: strings.TrimPrefix(r.URL.Path, "/"),
		SetCategoryRequest: utils.SetCategoryRequest{
			AppName:  appName,
			Category: utils.Category(category),
		},
	}
	res, err := a.commWithDaemonService(msg)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if !res.SetCategoryResponse.IsCategorySet {
		a.serverError(w, fmt.Errorf("error setting app category"))
		return
	}

	if err = views.SetCategory(category, appName).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}

func (a *App) getCategory(w http.ResponseWriter, r *http.Request) {
	appName := r.URL.Query().Get("name")
	if appName == "" {
		a.clientError(w, http.StatusBadRequest, errors.New("empty query param"))
		return
	}

	msg := utils.Message{Endpoint: strings.TrimPrefix(r.URL.Path, "/")}

	res, err := a.commWithDaemonService(msg)
	if err != nil {
		a.serverError(w, err)
		return
	}

	var appInfo utils.AppIconCategoryAndCmdLine
	for _, v := range lastAppInfos {
		if v.AppName == appName {
			appInfo = v
			break
		}
	}

	if err = views.CategoryModal(res.GetCategoryResponse, appInfo).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
}
