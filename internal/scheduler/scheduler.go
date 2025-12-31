package scheduler

import (
	"context"
	"core/internal/domain_model/model"
	"core/internal/notification"
	"core/internal/repo"
	"core/internal/scheduler/jobs"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

var cronEngine *cron.Cron
var deadlineJob *jobs.DeadlineReminderJob
var deadlineEntryID cron.EntryID

func Setup(ctx context.Context) {
	loc := time.Local
	cronEngine = cron.New(cron.WithLocation(loc))

	// Load scheduler config
	conf := model.SchedulerConfig{DailyTime: "09:00"}
	if setting, err := repo.GetSystemSetting(ctx, model.JOB_SCHEDULER); err == nil {
		if len(setting.Metadata) > 0 {
			_ = json.Unmarshal(setting.Metadata, &conf)
		}
	} else {
		slog.Debug("scheduler: using default schedule", "error", err)
	}

	// Build cron spec from DailyTime (HH:MM)
	spec := dailySpecFromTime(conf.DailyTime)

	// Register jobs 
	deadlineJob = jobs.NewDeadlineReminderJob(notification.NewService(notification.SystemEmailSender{}))
	entryID, err := cronEngine.AddFunc(spec, func() {
		jobCtx := context.Background()
		slog.Info("deadline job: TRIGGERED", "at", time.Now(), "loc", time.Now().Location().String())
		if err := deadlineJob.Run(jobCtx); err != nil {
			slog.Error("scheduler: deadline reminder job failed", "error", err)
		}
	})
	if err != nil {
		slog.Error("scheduler: failed to add deadline job", "error", err)
		return
	}
	deadlineEntryID = entryID

	cronEngine.Start()
	entry := cronEngine.Entry(entryID)
	slog.Info("scheduler: started", "spec", spec, "next_run", entry.Next, "prev_run", entry.Prev)
}

func Stop() {
	if cronEngine != nil {
		cronEngine.Stop()
	}
}

// dailySpecFromTime converts a HH:MM string into a 5-field cron spec (minute hour * * *).
// Defaults to 09:00 if parsing fails.
func dailySpecFromTime(hhmm string) string {
	def := "0 9 * * *"
	if hhmm == "" {
		return def
	}
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		return def
	}
	if h < 0 || h > 23 {
		h = 9
	}
	if m < 0 || m > 59 {
		m = 0
	}
	return fmt.Sprintf("%d %d * * *", m, h)
}

func RescheduleDaily(ctx context.Context, hhmm string) error {
	spec := dailySpecFromTime(hhmm)
	if cronEngine == nil {
		Setup(ctx)
	}

	// Remove existing entry if present
	if deadlineEntryID != 0 {
		cronEngine.Remove(deadlineEntryID)
		slog.Info("scheduler: removed previous deadline job entry", "entry_id", deadlineEntryID)
		deadlineEntryID = 0
	}

	if deadlineJob == nil {
		deadlineJob = jobs.NewDeadlineReminderJob(notification.NewService(notification.SystemEmailSender{}))
	}

	entryID, err := cronEngine.AddFunc(spec, func() {
		jobCtx := context.Background()
		slog.Info("deadline job: TRIGGERED", "at", time.Now(), "loc", time.Now().Location().String())
		if err := deadlineJob.Run(jobCtx); err != nil {
			slog.Error("scheduler: deadline reminder job failed", "error", err)
		}
	})
	if err != nil {
		slog.Error("scheduler: failed to reschedule deadline job", "error", err)
		return err
	}
	deadlineEntryID = entryID
	entry := cronEngine.Entry(entryID)
	slog.Info("scheduler: rescheduled", "spec", spec, "next_run", entry.Next, "prev_run", entry.Prev)
	return nil
}
