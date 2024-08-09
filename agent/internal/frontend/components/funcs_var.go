package frontend

import (
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"time"
)

var (
	art1 string = `
┓ •┏┓     ┳┳┓
┃ ┓┗┓┏┏┓┏┓┃┃┃┏┓┏┓
┗┛┗┗┛┗┛ ┗ ┛ ┗┗┛┛┗`

	art2 string = `
╦ 	┬ ╔═╗┌─┐┬─┐┌─┐╔╦╗┌─┐┌┐┌
║ 	│ ╚═╗│  ├┬┘├┤ ║║║│ ││││
╩═╝ ┴ ╚═╝└─┘┴└─└─┘╩ ╩└─┘┘└┘`

	art3 string = `

	LISCREMON
`
)

var arrayOfArt = []string{art1, art2, art3}

func getRandomArt() string {
	randonIndex := rand.Intn(len(arrayOfArt))
	return arrayOfArt[randonIndex]
}

func monthDropDownSelectArray(n int) []string {
	today := time.Now()
	firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	past3Month := make([]string, n)
	for i := 0; i < n; i++ {
		m := firstDayOfMonth.AddDate(0, -(i + 1), 0)
		past3Month[i] = m.Month().String()
	}
	return past3Month
}

func writeImageToFile(imageData []byte, filename string) bool {

	if len(imageData) == 0 {
		return false
	}
	var (
		tmpFolder = "/tmp/LiScreMon/"
		file      *os.File
		err       error
	)
	_, err = os.Stat(tmpFolder)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false
	}

	if err = os.MkdirAll(tmpFolder, 0755); err != nil {
		return false
	}

	if file, err = os.Create(tmpFolder + filename + ".png"); err != nil {
		return false
	}
	defer file.Close()

	if _, err = file.Write(imageData); err != nil {
		return false
	}

	return true
}

func getImageFilePath(filename string, imageData []byte) string {
	if _, err := os.Stat("/tmp/LiScreMon/" + filename + ".png"); err != nil {
		if writeImageToFile(imageData, filename) {
			return "/tmp/LiScreMon/" + filename + ".png"
		}
		return "/assets/image/noAppImage.jpg"
	}
	return "/tmp/LiScreMon/" + filename + ".png"
}

func formatTimeToHumanReadable(t time.Time) string {
	return t.Format("Monday, 02 January 2006, 03:04 PM")
}

func intToDuration(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}

func FormatTimeTOHMTLDatetimeLocal(t time.Time) string {
	return t.Format("2006-01-02T15:04")
}

func nextTwoWeeks() time.Time {
	return time.Now().AddDate(0, 0, 15)
}
