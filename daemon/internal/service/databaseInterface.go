package service

import (
	db "smDaemon/daemon/internal/database"

	"utils"

	"github.com/google/uuid"
)

type DatabaseInterface interface {
	GetDay(utils.Date) (db.DailyStat, error)
	GetWeek(utils.Date) (db.WeeklyStat, error)
	AppWeeklyStat(appName string, anyDayInTheWeek utils.Date) (utils.AppRangeStat, error)
	AppMonthlyStat(appName, month, year string) (utils.AppRangeStat, error)
	AppDateRangeStat(appName string, start, end utils.Date) (utils.AppRangeStat, error)
	GetAppIconCategoryAndCmdLine([]string) ([]utils.AppIconCategoryAndCmdLine, error)
	GetAllTask() ([]utils.Task, error)
	RemoveTask(id uuid.UUID) error
	SetAppCategory(appName string, category utils.Category) error
	GetAllACategories() ([]utils.Category, error)
	GetAllApp() ([]utils.AppIconCategoryAndCmdLine, error)
	GetTaskByUUID(taskID uuid.UUID) (utils.Task, error)
	GetAppTodayActiveStatSoFar(appName string) (float64, error)
}
