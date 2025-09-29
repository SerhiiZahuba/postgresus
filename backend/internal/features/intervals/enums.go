package intervals

type IntervalType string

const (
	IntervalHourly  IntervalType = "HOURLY"
	IntervalDaily   IntervalType = "DAILY"
	IntervalWeekly  IntervalType = "WEEKLY"
	IntervalMonthly IntervalType = "MONTHLY"
	InternalCron    IntervalType = "CRON"
)
