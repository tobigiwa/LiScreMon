package backend

import (
	"agent/internal/backend/chart"
	"context"
	"errors"
	"fmt"
	"net/http"

	"strings"
	"time"

	utils "utils"

	views "agent/internal/frontend/components"

	"github.com/a-h/templ"
)

var lastRequestSaturday string

func (a *App) WeekStatHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("week")
	endpoint := strings.TrimPrefix(r.URL.Path, "/")

	var (
		msg utils.Message
		err error
	)

	switch query {
	case "thisweek":
		msg = utils.Message{
			Endpoint:        endpoint,
			WeekStatRequest: utils.SaturdayOfTheWeek(time.Now()),
		}

	case "lastweek":
		msg = utils.Message{
			Endpoint:        endpoint,
			WeekStatRequest: utils.ToDateType(utils.PreviousWeekSaturday(time.Now())),
		}

	case "month":
		var firstSaturdayOfTheMonth utils.Date
		var q string
		if q = r.URL.Query().Get("month"); q == "" {
			a.clientError(w, http.StatusBadRequest, errors.New("query param:month: cannot be empty"))
			return
		}
		if firstSaturdayOfTheMonth = utils.FirstSaturdayOfTheMonth(q); firstSaturdayOfTheMonth == "" {
			a.clientError(w, http.StatusBadRequest, errors.New("query param:month: invalid data"))
			return
		}
		msg = utils.Message{
			Endpoint:        endpoint,
			WeekStatRequest: firstSaturdayOfTheMonth,
		}

	case "backward", "forward":
		var t time.Time
		if t, err = time.Parse(utils.TimeFormat, lastRequestSaturday); err != nil {
			a.clientError(w, http.StatusBadRequest, errors.New("header value 'lastSaturday' invalide"))
			return
		}

		if query == "backward" {
			msg = utils.Message{
				Endpoint:        endpoint,
				WeekStatRequest: utils.ToDateType(utils.PreviousWeekSaturday(t)),
			}
		}

		if query == "forward" {
			msg = utils.Message{
				Endpoint:        endpoint,
				WeekStatRequest: utils.ToDateType(utils.NexWeektSaturday(t)),
			}

			if utils.IsFutureDate(t) {
				msg = utils.Message{
					Endpoint:        endpoint,
					WeekStatRequest: utils.SaturdayOfTheWeek(time.Now()), // show current week
				}
			}
		}
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		a.serverError(w, fmt.Errorf("error occurred in commWithDaemonService:%w", err))
		return
	}

	if err = weekStatResponse(msg.WeekStatResponse).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
		return
	}

	lastRequestSaturday = msg.WeekStatResponse.Keys[6]
	clear(lastAppInfos)
	for _, v := range msg.WeekStatResponse.AppDetail {
		lastAppInfos = append(lastAppInfos, v.AppInfo)
	}
}

func weekStatResponse(w utils.WeekStatMessage) templ.Component {
	return views.WeekStatTempl(
		chart.WeekStatBarChart(chart.BarChartData{
			XAxis:       w.FormattedDay[:],
			YAxis:       w.Values[:],
			Keys:        w.Keys[:],
			Month:       w.Month,
			Year:        w.Year,
			TotalUptime: w.TotalWeekUptime,
		}),
		w.TotalWeekUptime,
		w.AppDetail,
		w.Keys[6],
	)
}
