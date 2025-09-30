package intervals

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Interval struct {
	ID       uuid.UUID    `json:"id"       gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Interval IntervalType `json:"interval" gorm:"type:text;not null"`

	TimeOfDay   *string `json:"timeOfDay"            gorm:"type:text;"`
	// only for WEEKLY
	Weekday     *int    `json:"weekday,omitempty"    gorm:"type:int"`
	// only for MONTHLY
	DayOfMonth  *int    `json:"dayOfMonth,omitempty" gorm:"type:int"`
	// only for CRON
	CronExpr    *string `json:"cronExpr,omitempty"   gorm:"type:text;"`
}

func (i *Interval) BeforeSave(tx *gorm.DB) error {
	return i.Validate()
}

func (i *Interval) Validate() error {
	// daily, weekly and monthly require timeOfDay
	if (i.Interval == IntervalDaily || i.Interval == IntervalWeekly || i.Interval == IntervalMonthly) &&
		i.TimeOfDay == nil {
		return errors.New("time of day is required for daily, weekly and monthly intervals")
	}

	if i.Interval == IntervalWeekly && i.Weekday == nil {
		return errors.New("weekday is required for weekly intervals")
	}

	if i.Interval == IntervalMonthly && i.DayOfMonth == nil {
		return errors.New("day of month is required for monthly intervals")
	}

	if i.Interval == IntervalCron {
		if i.CronExpr == nil || *i.CronExpr == "" {
			return errors.New("cron expression is required for CRON intervals")
		}
		// validate cron expr
		_, err := cron.ParseStandard(*i.CronExpr)
		if err != nil {
			return errors.New("invalid cron expression: " + err.Error())
		}
	}

	return nil
}

func (i *Interval) ShouldTriggerBackup(now time.Time, lastBackupTime *time.Time) bool {
	if lastBackupTime == nil {
		return true
	}

	switch i.Interval {
	case IntervalHourly:
		return now.Sub(*lastBackupTime) >= time.Hour
	case IntervalDaily:
		return i.shouldTriggerDaily(now, *lastBackupTime)
	case IntervalWeekly:
		return i.shouldTriggerWeekly(now, *lastBackupTime)
	case IntervalMonthly:
		return i.shouldTriggerMonthly(now, *lastBackupTime)
	case IntervalCron:
		return i.shouldTriggerCron(now, *lastBackupTime)
	default:
		return false
	}
}


func (i *Interval) shouldTriggerDaily(now, lastBackup time.Time) bool {
	return now.Sub(lastBackup) >= 24*time.Hour
}

func (i *Interval) shouldTriggerWeekly(now, lastBackup time.Time) bool {
	return now.Sub(lastBackup) >= 7*24*time.Hour
}

func (i *Interval) shouldTriggerMonthly(now, lastBackup time.Time) bool {
	return now.Sub(lastBackup) >= 30*24*time.Hour
}


// CRON trigger
func (i *Interval) shouldTriggerCron(now, lastBackup time.Time) bool {
	if i.CronExpr == nil {
		return false
	}
	sched, err := cron.ParseStandard(*i.CronExpr)
	if err != nil {
		return false
	}

	// find next run after lastBackup
	next := sched.Next(lastBackup)
	
	return now.After(next) || now.Equal(next)
}
