package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

func EncodeJSON[T any](tyPe T) ([]byte, error) {
	encoded, err := json.Marshal(tyPe)
	if err != nil {
		return nil, fmt.Errorf("%v:%w", err, ErrSerialization)
	}
	return encoded, nil
}

func DecodeJSON[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("%v:%w", err, ErrDeserialization)
	}
	return result, nil
}

func HrsAndMinute(hr float64) (int, int) {
	return int(hr), int(math.Round((hr - float64(int(hr))) * 60))
}

func UsageTimeInHrsMin(f float64) string {
	hrs, min := HrsAndMinute(f)
	return fmt.Sprintf("%dHrs:%dMin", hrs, min)
}

func AddOrdinalSuffix(n int) string {
	switch n {
	case 1, 21, 31:
		return fmt.Sprintf("%dst", n)
	case 2, 22:
		return fmt.Sprintf("%dnd", n)
	case 3, 23:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

func ToTimeType(d Date) (t time.Time) {
	t, _ = time.Parse(TimeFormat, string(d)) // This is safe because the date is already validated
	return
}

func ToDateType(t time.Time) Date {
	return Date(t.Format(TimeFormat))
}

func IsFutureWeek(t time.Time) bool {
	now := time.Now()

	// Adjust the week start day to Sunday
	now = now.AddDate(0, 0, int(time.Sunday-now.Weekday()))
	t = t.AddDate(0, 0, int(time.Sunday-t.Weekday()))

	_, currentWeek := now.ISOWeek()
	_, inputWeek := t.ISOWeek()

	return inputWeek > currentWeek
}

func IsPastWeek(t time.Time) bool {
	now := time.Now()

	// Adjust the week start day to Sunday
	now = now.AddDate(0, 0, int(time.Sunday-now.Weekday()))
	t = t.AddDate(0, 0, int(time.Sunday-t.Weekday()))

	_, currentWeek := now.ISOWeek()
	_, inputWeek := t.ISOWeek()

	return inputWeek < currentWeek
}

func DaysInThatWeek(w time.Time) [7]Date {
	var arr [7]Date
	startOftheWeek := w.AddDate(0, 0, -int(w.Weekday()))
	for i := 0; i < 7; i++ {
		arr[i] = Date(fmt.Sprint(startOftheWeek.AddDate(0, 0, i).Format(TimeFormat)))
	}
	return arr
}
func Today() Date {
	return Date(time.Now().Format(TimeFormat))
}
func FormattedDay(date Date) string {
	day := ToTimeType(date)
	dayWithSuffix := AddOrdinalSuffix(day.Day())
	weekDay := strings.TrimSuffix(day.Weekday().String(), "day")
	return fmt.Sprintf("%v. %v", weekDay, dayWithSuffix)
}

func MonthAndYear(date Date) (month, year string) {
	y, mon, _ := ToTimeType(date).Date()
	month, year = mon.String(), fmt.Sprint(y)
	return month, year
}

func calculatePercentage(part, total float64) float64 {
	return (part / total) * 100
}

func PercentagesString(part, total float64) string {
	p := calculatePercentage(part, total)
	if math.IsNaN(p) {
		return "NaN"
	}
	return fmt.Sprintf("%.0f%%", p)
}

func SaturdayOfTheWeek(t time.Time) Date {
	daysUntilSaturday := 6 - int(t.Weekday())
	return Date(t.AddDate(0, 0, daysUntilSaturday).Format(TimeFormat))
}

func FirstSaturdayOfTheMonth(month string) Date {
	t, err := time.Parse("January", month)
	if err != nil {
		return ""
	}
	year := time.Now().Year()
	firstDayOfMonth := time.Date(year, t.Month(), 1, 0, 0, 0, 0, time.UTC)

	for {
		if firstDayOfMonth.Weekday() == time.Saturday {
			return Date(firstDayOfMonth.Format(TimeFormat))
		}
		firstDayOfMonth = firstDayOfMonth.AddDate(0, 0, 1)
	}
}

func LastSaturdayOfTheMonth(month string) string {
	t, err := time.Parse("January", month)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	NextMonth := time.Date(time.Now().Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	var s time.Time
	for {
		NextMonth = NextMonth.AddDate(0, 0, -1)
		if NextMonth.Weekday() == time.Saturday {
			s = NextMonth
			break
		}
	}
	return s.Format(TimeFormat)
}

func PreviousWeekSaturday(t time.Time) time.Time {

	if t.Weekday() == time.Saturday {
		return t.AddDate(0, 0, -7)
	}

	daysSinceSaturday := (t.Weekday() + 1) % 7
	return t.AddDate(0, 0, -int(daysSinceSaturday))
}

func NexWeektSaturday(t time.Time) time.Time {

	if t.Weekday() == time.Saturday {
		return t.AddDate(0, 0, 7)
	}

	daysSinceSaturday := (t.Weekday() + 1) % 7
	return t.AddDate(0, 0, int(daysSinceSaturday))
}
func IsFutureDate(t time.Time) bool {
	today := time.Now()
	nextWeekDay := t.AddDate(0, 0, 7)
	return nextWeekDay.After(today)
}

func FormattedToDay() time.Time {
	return ToTimeType(Date(time.Now().Format(TimeFormat)))
}

func AllTheDaysInMonth(year, month string) ([]Date, error) {
	t, err := time.Parse("2006 January", year+" "+month)
	if err != nil {
		return nil, fmt.Errorf("parse %w", err)
	}

	fmt.Println(t.Day(), t.Month(), t.Year())
	lastDayOfTheGivenMonth := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()

	dates := make([]Date, 0, lastDayOfTheGivenMonth)

	for day := 1; day <= lastDayOfTheGivenMonth; day++ {
		dates = append(dates, Date(time.Date(t.Year(), t.Month(), day, 0, 0, 0, 0, t.Location()).Format(TimeFormat)))
	}

	return dates, nil
}

func ValidDateType(s string) bool {
	return DateTypeRegexPattern.MatchString(s)
}

func IsInCurrentWeekTime(t time.Time) bool {
	now := time.Now()
	_, currentWeek := now.ISOWeek()
	_, tWeek := t.ISOWeek()

	return currentWeek == tWeek && now.Year() == t.Year()
}

func IsInCurrentWeekDate(d Date) bool {
	return SaturdayOfTheWeek(time.Now()) == d
}

func NotifyWithBeep(title, subtitle string) {
	beeep.Alert(title, subtitle, APP_LOGO_FILE_PATH)
}
func NotifyWithoutBeep(title, subtitle string) {
	beeep.Notify(title, subtitle, APP_LOGO_FILE_PATH)
}
