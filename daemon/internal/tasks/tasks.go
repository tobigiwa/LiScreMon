// Package tasks is meant to handle functionalities that
// would need scheduling, the github.com/go-co-op/gocron/v2
// library is used for that.
// NOTE: This package already depends on `package screen`.
package tasks

import (
	"fmt"
	"log"
	"os/exec"
	monitoring "smDaemon/daemon/internal/screen/linux"

	"reflect"
	"time"

	utils "utils"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type TaskManagerDbRequirement interface {
	GetAllTask() ([]utils.Task, error)
	RemoveTask(id uuid.UUID) error
	AddTask(task utils.Task) error
	GetAppTodayActiveStatSoFar(appName string) (float64, error)
	ReportWeeklyUsage(anyDayInTheWeek time.Time) (string, error)
}

type TaskManager struct {
	dbHandle TaskManagerDbRequirement
	gocron   gocron.Scheduler
	channel  chan utils.Task
}

func newTaskManger(dbHandle TaskManagerDbRequirement) *TaskManager {
	var tm TaskManager
	tm.dbHandle = dbHandle
	tm.gocron, _ = gocron.NewScheduler()
	tm.channel = make(chan utils.Task)
	return &tm
}

func (tm *TaskManager) CloseChan() error {
	if err := tm.gocron.Shutdown(); err != nil {
		log.Println("error shutting down gocron Scheduler:", err)
		return err
	}
	tm.channel <- utils.Task{}
	return nil
}

func (tm *TaskManager) SendTaskToTaskManager(task utils.Task) error {
	if err := tm.dbHandle.AddTask(task); err != nil {
		return fmt.Errorf("error adding task to db: %w", err)
	}

	if reflect.ValueOf(task).IsZero() {
		return utils.ErrZeroValueTask
	}

	tm.channel <- task
	return nil
}

func StartTaskManger(dbHandle TaskManagerDbRequirement) (*TaskManager, error) {

	tm := newTaskManger(dbHandle)

	tasks, err := tm.dbHandle.GetAllTask()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", utils.ErrTaskMangerNotStarted.Error(), err)
	}

	go tm.disperseTask()

	for _, task := range tasks {

		switch {
		case task.Job != utils.DailyAppLimit: // i.e reminders
			if task.Reminder.StartTime.Before(time.Now()) { // this is a double-check, the reminder should have been removed when done
				if err := tm.dbHandle.RemoveTask(task.UUID); err != nil {
					return nil, fmt.Errorf("%s %+v :err %w", utils.ErrDeletingTask.Error(), task, err)
				}
			}

		case task.Job == utils.DailyAppLimit:
			if task.AppLimit.OneTime && task.AppLimit.Day != utils.Today() { // this is a double-check, the oneTime app limit should have been removed when done;
				if err := tm.dbHandle.RemoveTask(task.UUID); err != nil { // whether it reached the limit or not,for a new day, the limit does not matter again, hence why it is a dailyLimit.
					return nil, fmt.Errorf("%s: %+v :%w", utils.ErrDeletingTask.Error(), task, err)
				}
			}
		}

		tm.channel <- task
	}

	if _, err := tm.gocron.NewJob(
		gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(time.Sunday),
			gocron.NewAtTimes(
				gocron.NewAtTime(9, 0, 0),
			),
		),
		gocron.NewTask(
			func() {
				s, err := tm.dbHandle.ReportWeeklyUsage(utils.PreviousWeekSaturday(time.Now()))
				if err != nil {
					log.Println("reportlyweek", err)
					return
				}

				utils.NotifyWithBeep("Weekly Screentime", s)
			},
		),
	); err != nil {
		log.Println("error creating weekly job", err)
	}

	return tm, nil
}

func (tm *TaskManager) disperseTask() {

	tm.gocron.Start()

	for {
		task := <-tm.channel

		if reflect.ValueOf(task).IsZero() {
			close(tm.channel)
			log.Println("closing and cleaning TaskManager")
			break
		}

		fmt.Printf("task received   %+v\n\n", task)

		switch task.Job {
		case utils.ReminderWithNoAppLaunch:
			tm.reminderWithNoAppLaunch(task)

		case utils.ReminderWithAppLaunch:
			tm.reminderWithAppLaunch(task)

		case utils.DailyAppLimit:
			timeSofar, _ := tm.dbHandle.GetAppTodayActiveStatSoFar(task.AppName)
			monitoring.AddNewLimit(task, timeSofar)
		}
	}
}

func (tm *TaskManager) reminderWithNoAppLaunch(task utils.Task) {

	tm.preNotify(task)

	if _, err := tm.gocron.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(task.Reminder.StartTime)),
		gocron.NewTask(reminderAlert, task.UI, true),
		gocron.WithTags(task.UUID.String()),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				tm.dbHandle.RemoveTask(task.UUID)

			}),
		),
	); err != nil {
		log.Println("error creating job", err)
	}

}

func (tm *TaskManager) reminderWithAppLaunch(task utils.Task) {
	tm.preNotify(task)

	tm.gocron.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(task.Reminder.StartTime)),
		gocron.NewTask(reminderAlert, task.UI, true),
		gocron.WithTags(task.UUID.String()),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					cmd := exec.Command("bash", "-c", task.CmdLine)
					err := cmd.Start()
					if err != nil {
						log.Println(err)
					}

					err = cmd.Wait()
					if err != nil {
						log.Printf("Command finished with error: %v", err)
					}
				}),
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				tm.dbHandle.RemoveTask(task.UUID)
			}),
		),
	)
}

func (tm *TaskManager) preNotify(task utils.Task) {

	for i := 0; i < 2; i++ {
		var t time.Time
		notifyBeForeReminder, withSound := task.Reminder.AlertTimesInMinutes[i], task.Reminder.AlertSound[i]

		if t = task.Reminder.StartTime.Add(-time.Duration(notifyBeForeReminder) * time.Minute); t.Before(time.Now()) {
			continue // reminder is the past, useless
		}

		if _, err := tm.gocron.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
			gocron.NewTask(preNotifyAlert, task.UI.Title, notifyBeForeReminder, withSound),
			gocron.WithTags(task.UUID.String())); err != nil {
			log.Println("gocron failed to add notififcation", err)
		}
	}
}

func preNotifyAlert(taskTitle string, durationbeforeTask int, withSound bool) {
	title := fmt.Sprintf("%d minutes to your task", durationbeforeTask)

	if withSound {
		utils.NotifyWithBeep(title, taskTitle)
		return
	}

	utils.NotifyWithoutBeep(title, taskTitle)
}

func reminderAlert(task utils.UItextInfo, withSound bool) {
	title := fmt.Sprintf("Task: %s", task.Title)

	if withSound {
		utils.NotifyWithBeep(title, task.Subtitle)
		return
	}

	utils.NotifyWithoutBeep(title, task.Subtitle)
}

func (tm *TaskManager) RemoveTask(taskUUID uuid.UUID) {
	tm.gocron.RemoveByTags(taskUUID.String())
}
