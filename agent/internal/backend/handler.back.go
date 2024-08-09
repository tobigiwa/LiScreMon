package backend

import (
	"agent/internal/backend/chart"
	views "agent/internal/frontend/components"
	"context"
	"fmt"
	"net/http"
	"strings"
	utils "utils"
)

func (a *App) backHandler(w http.ResponseWriter, r *http.Request) {
	var (
		msg utils.Message
		err error
	)
	msg = utils.Message{
		Endpoint:        strings.TrimPrefix(r.URL.Path, "/"),
		WeekStatRequest: utils.Date(lastRequestSaturday),
	}

	if msg, err = a.commWithDaemonService(msg); err != nil {
		a.serverError(w, fmt.Errorf("error occurred in commWithDaemonService:%w", err))
		return
	}

	chart := chart.WeekStatBarChart(chart.BarChartData{
		XAxis:       msg.WeekStatResponse.FormattedDay[:],
		YAxis:       msg.WeekStatResponse.Values[:],
		Keys:        msg.WeekStatResponse.Keys[:],
		Month:       msg.WeekStatResponse.Month,
		Year:        msg.WeekStatResponse.Year,
		TotalUptime: msg.WeekStatResponse.TotalWeekUptime,
	})

	if err = views.ChartOnly(chart).Render(context.TODO(), w); err != nil {
		a.serverError(w, err)
		return
	}
}
