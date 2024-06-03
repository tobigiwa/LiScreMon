package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"pkg/types"
	"strings"
	"time"

	helperFuncs "pkg/helper"
)

var lastRequestSaturday string

func (a *App) WeekStatHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("week")
	endpoint := strings.TrimPrefix(r.URL.Path, "/")

	var (
		msg types.Message
		err error
	)

	switch query {
	case "thisweek":
		msg = types.Message{
			Endpoint:        endpoint,
			WeekStatRequest: helperFuncs.SaturdayOfTheWeek(time.Now()),
		}

	case "lastweek":
		msg = types.Message{
			Endpoint:        endpoint,
			WeekStatRequest: helperFuncs.ReturnLastWeekSaturday(time.Now()),
		}

	case "month":
		var firstSaturdayOfTheMonth, q string
		if q = r.URL.Query().Get("month"); q == "" {
			a.clientError(w, http.StatusBadRequest, errors.New("query param:month: cannot be empty"))
			return
		}
		if firstSaturdayOfTheMonth = helperFuncs.FirstSaturdayOfTheMonth(q); firstSaturdayOfTheMonth == "" {
			a.clientError(w, http.StatusBadRequest, errors.New("query param:month: invalid data"))
			return
		}
		msg = types.Message{
			Endpoint:        endpoint,
			WeekStatRequest: firstSaturdayOfTheMonth,
		}

	case "backward", "forward":
		var t time.Time
		if t, err = time.Parse(types.TimeFormat, lastRequestSaturday); err != nil {
			a.clientError(w, http.StatusBadRequest, errors.New("header value 'lastSaturday' invalide"))
			return
		}

		if query == "backward" {
			msg = types.Message{
				Endpoint:        endpoint,
				WeekStatRequest: helperFuncs.ReturnLastWeekSaturday(t),
			}
		}

		if query == "forward" {
			msg = types.Message{
				Endpoint:        endpoint,
				WeekStatRequest: helperFuncs.ReturnNexWeektSaturday(t),
			}

			if helperFuncs.IsFutureDate(t) {
				msg = types.Message{
					Endpoint:        endpoint,
					WeekStatRequest: helperFuncs.SaturdayOfTheWeek(time.Now()), // show current week
				}
			}
		}
	}

	if msg, err = a.writeAndReadWithDaemonService(msg); err != nil {
		a.serverError(w, fmt.Errorf("error occurred in writeAndReadWithDaemonService:%w", err))
		return
	}

	templComp := prepareHtTMLResponse(msg)

	if err = templComp.Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
	}
	lastRequestSaturday = msg.WeekStatResponse.Keys[6]
}