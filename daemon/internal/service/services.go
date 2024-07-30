package service

import (
	"fmt"
	db "smDaemon/daemon/internal/database"
	monitoring "smDaemon/daemon/internal/screen/linux"
	"smDaemon/daemon/internal/tasks"

	"cmp"

	"slices"
	"strings"
	"time"
	utils "utils"
)

type Service struct {
	db          DatabaseInterface
	taskManager *tasks.TaskManager
}

func (s *Service) StopTaskManger() error {
	return s.taskManager.CloseChan()
}

func (s *Service) getWeekStat(msg utils.Date) (utils.WeekStatMessage, error) {
	var (
		weekStat    db.WeeklyStat
		appsInfo    []utils.AppIconCategoryAndCmdLine
		allCategory []utils.Category
		err         error
	)

	if weekStat, err = s.db.GetWeek(msg); err != nil {
		return utils.NoMessage.WeekStatResponse, fmt.Errorf("error weekStat: %w", err)
	}

	var (
		keys             = [7]string{}
		formattedDay     = [7]string{}
		values           = [7]float64{}
		sizeOfApps       = len(weekStat.EachApp)
		appNameInTheWeek = make([]string, 0, sizeOfApps)
		appCard          = make([]utils.ApplicationDetail, 0, sizeOfApps)
	)

	for i := 0; i < 7; i++ {
		keys[i] = string(weekStat.DayByDayTotal[i].Key)
		values[i] = weekStat.DayByDayTotal[i].Value.Active
		formattedDay[i] = utils.FormattedDay(utils.Date(weekStat.DayByDayTotal[i].Key))
	}

	month, year := utils.MonthAndYear(utils.Date(weekStat.DayByDayTotal[6].Key))

	for i := 0; i < sizeOfApps; i++ {
		appNameInTheWeek = append(appNameInTheWeek, weekStat.EachApp[i].AppName)
	}

	if appsInfo, err = s.db.GetAppIconCategoryAndCmdLine(appNameInTheWeek); err != nil {
		return utils.NoMessage.WeekStatResponse, fmt.Errorf("err with GetAppIconAndCategory:%w", err)
	}

	for i := 0; i < sizeOfApps; i++ {
		appCard = append(appCard, utils.ApplicationDetail{AppInfo: appsInfo[i], Usage: weekStat.EachApp[i].Usage.Active})
	}

	if allCategory, err = s.db.GetAllACategories(); err != nil {
		return utils.NoMessage.WeekStatResponse, fmt.Errorf("err with GetAllCategories:%w", err)
	}

	return utils.WeekStatMessage{
			Keys:            keys,
			FormattedDay:    formattedDay,
			Values:          values,
			TotalWeekUptime: weekStat.WeekTotal.Active,
			AllCategory:     allCategory,
			Month:           month,
			Year:            fmt.Sprint(year),
			AppDetail:       appCard,
		},
		nil
}

func (s *Service) getAppStat(msg utils.AppStatRequest) (utils.AppStatMessage, error) {
	var (
		appStat utils.AppRangeStat
		err     error
	)

	switch msg.StatRange {
	case "week":
		appStat, err = s.db.AppWeeklyStat(msg.AppName, msg.Start)
	case "month":
		appStat, err = s.db.AppMonthlyStat(msg.AppName, msg.Month, msg.Year)
	case "range":
		appStat, err = s.db.AppDateRangeStat(msg.AppName, msg.Start, msg.End)
	}

	if err != nil {
		return utils.NoMessage.AppStatResponse, err
	}

	var (
		formattedDay      = make([]string, 0, len(appStat.DaysRange))
		values            = []float64{}
		lastDayOfTheRange = len(appStat.DaysRange) - 1
	)

	for i := 0; i < len(appStat.DaysRange); i++ {
		formattedDay = append(formattedDay, utils.FormattedDay(utils.Date(appStat.DaysRange[i].Key)))
		values = append(values, appStat.DaysRange[i].Value.Active)
	}
	month, year := utils.MonthAndYear(utils.Date(appStat.DaysRange[lastDayOfTheRange].Key))

	return utils.AppStatMessage{
			FormattedDay:     formattedDay,
			Values:           values,
			Month:            month,
			Year:             year,
			TotalRangeUptime: appStat.TotalRange.Active,
			AppInfo:          appStat.AppInfo,
		},
		nil
}

func (s *Service) getDayStat(msg utils.Date) (utils.DayStatMessage, error) {
	dayStat, err := s.db.GetDay(msg)
	if err != nil {
		return utils.NoMessage.DayStatResponse, err
	}
	d := utils.ToTimeType(msg)
	date := fmt.Sprintf("%s. %s %s, %d", strings.TrimSuffix(d.Weekday().String(), "day"), utils.AddOrdinalSuffix(d.Day()), d.Month().String(), d.Year())

	return utils.DayStatMessage{EachApp: dayStat.EachApp, DayTotal: dayStat.DayTotal, Date: date}, nil
}

func (s *Service) setAppCategory(msg utils.SetCategoryRequest) (utils.SetCategoryResponse, error) {
	if err := s.db.SetAppCategory(msg.AppName, msg.Category); err != nil {
		return utils.NoMessage.SetCategoryResponse, err
	}
	return utils.SetCategoryResponse{IsCategorySet: true}, nil
}

func (s *Service) tasks() (utils.ReminderMessage, error) {

	allApps, err := s.db.GetAllApp()
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}
	return utils.ReminderMessage{AllApps: allApps}, nil
}

func (s *Service) allReminderTask() (utils.ReminderMessage, error) {

	tasks, err := s.db.GetAllTask()
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}

	validTask := make([]utils.Task, 0, len(tasks))
	for _, task := range tasks {

		if task.Job == utils.DailyAppLimit {
			continue
		}

		now, taskStartTime := time.Now(), task.Reminder.StartTime

		if taskStartTime.Before(now) {
			if err := s.db.RemoveTask(task.UUID); err != nil {
				return utils.NoMessage.ReminderAndLimitResponse, err
			}
		}
		validTask = append(validTask, task)
	}

	slices.SortFunc(validTask, func(a, b utils.Task) int {
		return a.Reminder.StartTime.Compare(b.Reminder.StartTime)
	})

	return utils.ReminderMessage{AllTask: slices.Clip(validTask)}, nil
}

func (s *Service) allDailyAppLimitTask() (utils.ReminderMessage, error) {
	tasks, err := s.db.GetAllTask()
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}

	limitTask := make([]utils.Task, 0, len(tasks))

	for _, task := range tasks {
		if task.Job == utils.DailyAppLimit {
			limitTask = append(limitTask, task)
		}
	}

	slices.SortFunc(limitTask, func(a, b utils.Task) int {
		return cmp.Compare(a.AppLimit.Limit, b.AppLimit.Limit)
	})

	return utils.ReminderMessage{AllTask: slices.Clip(limitTask)}, nil
}

func (s *Service) addNewReminder(task utils.Task) (utils.ReminderMessage, error) {

	if task.Job == utils.ReminderWithAppLaunch {
		appInfo, err := s.db.GetAppIconCategoryAndCmdLine([]string{task.AppName})
		if err != nil {
			return utils.NoMessage.ReminderAndLimitResponse, err
		}
		task.AppIconCategoryAndCmdLine = appInfo[0]
	}

	err := s.taskManager.SendTaskToTaskManager(task)
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}

	return utils.ReminderMessage{TaskOptSuccessful: true}, nil
}

func (s *Service) addNewLimitApp(msg utils.Task) (utils.ReminderMessage, error) {

	err := s.taskManager.SendTaskToTaskManager(msg)
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}

	t, err := s.allDailyAppLimitTask()
	if err != nil {
		return utils.ReminderMessage{TaskOptSuccessful: true}, err
	}

	t.TaskOptSuccessful = true
	return t, nil
}

func (s *Service) removeTask(msg utils.Task) (utils.ReminderMessage, error) {

	if t, err := s.db.GetTaskByUUID(msg.UUID); err == nil {

		if t.Job == utils.DailyAppLimit {
			delete(monitoring.LimitApp, t.AppName)
		}

		s.taskManager.RemoveTask(msg.UUID)
	}

	err := s.db.RemoveTask(msg.UUID)
	if err != nil {
		return utils.NoMessage.ReminderAndLimitResponse, err
	}

	return utils.ReminderMessage{TaskOptSuccessful: true}, nil
}
