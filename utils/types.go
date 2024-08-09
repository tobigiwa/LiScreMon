package utils

import (
	"errors"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/google/uuid"
)

type Message struct {
	Endpoint            string              `json:"endpoint"`
	StatusCheck         string              `json:"statusCheck"`
	IsError             bool                `json:"isError"`
	Error               string              `json:"error"`
	DayStatRequest      Date                `json:"dayStatRequest"`
	DayStatResponse     DayStatMessage      `json:"dayStatResponse"`
	WeekStatRequest     Date                `json:"weekStatRequest"`
	WeekStatResponse    WeekStatMessage     `json:"weekStatResponse"`
	AppStatRequest      AppStatRequest      `json:"appStatResquest"`
	AppStatResponse     AppStatMessage      `json:"appStatResponse"`
	TaskRequest         Task                `json:"taskRequest"`
	TaskResponse        TaskMessage         `json:"taskResponse"`
	SetCategoryRequest  SetCategoryRequest  `json:"setCategoryRequest"`
	SetCategoryResponse SetCategoryResponse `json:"setCategoryResponse"`
	GetCategoryResponse []Category          `json:"getCategoryResponse"`
}

type SetCategoryRequest struct {
	AppName  string   `json:"appName"`
	Category Category `json:"category"`
}

type SetCategoryResponse struct {
	IsCategorySet bool `json:"isCategorySet"`
}

type TaskMessage struct {
	Task              Task                        `json:"task"`
	TaskOptSuccessful bool                        `json:"taskOptsuccessful"`
	AllTask           []Task                      `json:"allTask"`
	AllApps           []AppIconCategoryAndCmdLine `json:"allApps"`
}

type WeekStatMessage struct {
	Keys            [7]string           `json:"keys"`
	FormattedDay    [7]string           `json:"formattedDay"`
	Values          [7]float64          `json:"values"`
	TotalWeekUptime float64             `json:"totalWeekUptime"`
	Month           string              `json:"month"`
	Year            string              `json:"year"`
	AppDetail       []ApplicationDetail `json:"appDetail"`
}

type DayStatMessage struct {
	EachApp  []AppStat `json:"eachApp"`
	DayTotal Stats     `json:"dayTotal"`
	Date     string    `josn:"date"`
}

type AppStatRequest struct {
	AppName   string `json:"appName"`
	Month     string `json:"month"`
	Year      string `json:"year"`
	StatRange string `json:"statRange"`
	Start     Date   `json:"start"`
	End       Date   `json:"end"`
}

type AppStatMessage struct {
	FormattedDay     []string                  `json:"formattedDay"`
	Values           []float64                 `json:"values"`
	Month            string                    `json:"month"`
	Year             string                    `json:"year"`
	TotalRangeUptime float64                   `json:"totalRangeUptime"`
	AppInfo          AppIconCategoryAndCmdLine `json:"appInfo"`
}

type AppIconCategoryAndCmdLine struct {
	AppName           string   `json:"appName"`
	Icon              []byte   `json:"icon"`
	IsIconSet         bool     `json:"isIconSet"`
	IsCmdLineSet      bool     `json:"isCmdLine"`
	CmdLine           string   `json:"cmdLine"`
	Category          Category `json:"category"`
	IsCategorySet     bool     `json:"isCategorySet"`
	DesktopCategories []string `json:"desktopCategories"`
}

type ApplicationDetail struct {
	AppInfo      AppIconCategoryAndCmdLine `json:"appInfo"`
	Usage        float64                   `json:"usage"`
	AnyDayInStat Date                      `json:"anyDayInStat"`
}

type GenericKeyValue[K, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

type AppRangeStat struct {
	AppInfo    AppIconCategoryAndCmdLine      `json:"appInfo"`
	DaysRange  []GenericKeyValue[Date, Stats] `json:"daysRange"`
	TotalRange Stats                          `json:"totalRange"`
}

type AppStat struct {
	AppName string `json:"appName"`
	Usage   Stats  `json:"usage"`
}

type (

	// date underneath is a
	/* string of a time.Time format. "2006-01-02" */
	Date       string
	ScreenType string
	Category   string
)

func (d Date) ToTime() (time.Time, error) {
	if !DateTypeRegexPattern.MatchString(string(d)) {
		return time.Time{}, errors.New("invalid date format")
	}
	return time.Parse(TimeFormat, string(d))
}
func (c Category) String() string {
	return string(c)
}

type Stats struct {
	Active         float64        `json:"active"`
	Open           float64        `json:"open"`
	Inactive       float64        `json:"inactive"`
	ActiveTimeData []TimeInterval `json:"activeTimeData"`
}

type TimeInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ScreenTime struct {
	WindowID xproto.Window `json:"windowID"`
	AppName  string        `json:"appName"`
	Type     ScreenType    `json:"type"`
	Duration float64       `json:"duration"`
	Interval TimeInterval  `json:"interval"`
}

type Task struct {
	AppIconCategoryAndCmdLine
	TaskTime
	UUID uuid.UUID  `json:"uuid"`
	UI   UItextInfo `json:"ui"`
	Job  TaskType   `json:"job"`
}

type UItextInfo struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Notes    string `json:"notes"`
}
type TaskType string

type TaskTime struct {
	AppLimit AppLimit `json:"appLimit"`
	Reminder Reminder `json:"reminder"`
}

type Reminder struct {
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	AlertTimesInMinutes [3]int    `json:"alertTimesInMinutes"`
	AlertSound          [3]bool   `json:"alertSound"`
}

type AppLimit struct {
	Limit          float64 `json:"limit"`
	OneTime        bool    `json:"oneTime"`
	ExitApp        bool    `json:"exitApp"`
	Day            Date    `json:"today"`
	IsLimitReached bool    `json:"isLimitReached"`
}

type ConfigFile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	BrowserAddr string `json:"browserAddr"`
}
