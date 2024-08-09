package utils

import (
	"fmt"
	"regexp"

	"github.com/BurntSushi/xgbutil"
)

var X11Connection *xgbutil.XUtil

var (
	ErrDeserialization = fmt.Errorf("error deserializing data")
	ErrSerialization   = fmt.Errorf("error serializing data")
	ErrLimitAppExist   = fmt.Errorf("limitApp task already exist")

	ErrDeletingTask         = fmt.Errorf("err deleting old task")
	ErrTaskMangerNotStarted = fmt.Errorf("taskManager could not be started")

	ErrZeroValueTask = fmt.Errorf("sent task cannot be empty struct")
)

var (
	NoMessage                   = Message{}
	NoAppIconCategoryAndCmdLine = AppIconCategoryAndCmdLine{}
	InvalidDateType             = Date("")
	DateTypeRegexPattern        = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

const (
	Active     ScreenType = "active"
	Inactive   ScreenType = "inactive"
	Open       ScreenType = "open"
	TimeFormat string     = "2006-01-02"

	ReminderWithNoAppLaunch TaskType = "ReminderWithNoAppLaunch"
	ReminderWithAppLaunch   TaskType = "ReminderWithAppLaunch"
	DailyAppLimit           TaskType = "DailyAppLimit"
)

const (
	ProductivityAndUtility           Category = "Productivity & Utility"
	CommunicationAndSocialNetworking Category = "Communication & Social Networking"
	EntertainmentAndGaming           Category = "Entertainment & Gaming"
	WebBrowser                       Category = "Web Browser"
	SytemApp                         Category = "System App"
)

var CategoryMap map[string]Category = map[string]Category{
	"utilities":        ProductivityAndUtility,
	"productivity":     ProductivityAndUtility,
	"texteditor":       ProductivityAndUtility,
	"development":      ProductivityAndUtility,
	"ide":              ProductivityAndUtility,
	"editor":           ProductivityAndUtility,
	"viewer":           ProductivityAndUtility,
	"office":           ProductivityAndUtility,
	"communication":    CommunicationAndSocialNetworking,
	"social":           CommunicationAndSocialNetworking,
	"instantmessaging": CommunicationAndSocialNetworking,
	"entertainment":    EntertainmentAndGaming,
	"gaming":           EntertainmentAndGaming,
	"player":           EntertainmentAndGaming,
	"video":            EntertainmentAndGaming,
	"audio":            EntertainmentAndGaming,
	"audiovideo":       EntertainmentAndGaming,
	"browser":          WebBrowser,
	"webbrowser":       WebBrowser,
	"gnome":            SytemApp,
	"terminal":         SytemApp,
	"system":           SytemApp,
	"filemanager":      SytemApp,
	"filesystem":       SytemApp,
	"core":             SytemApp,
	"packagemanager":   SytemApp,
	"settings":         SytemApp,
	"terminalemulator": SytemApp,
}

var DefalutCategory = []Category{ProductivityAndUtility, CommunicationAndSocialNetworking,
	EntertainmentAndGaming, WebBrowser, SytemApp}
