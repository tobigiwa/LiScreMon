package backend

import (
	"fmt"
	"net/http"
)

var (
	ErrLinkExpired  = fmt.Errorf("link expired")
	ErrDuplicateKey = fmt.Errorf("user with email account already exit")
)

func (a App) clientError(w http.ResponseWriter, errStatus int, err error) {
	http.Error(w, err.Error(), errStatus)
	a.logger.Error("clientError" + err.Error())
}

func (a App) serverError(w http.ResponseWriter, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	a.logger.Error("serverError" + err.Error())
}
