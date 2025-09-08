package notifiers

type NotifierType string

const (
	NotifierTypeEmail    NotifierType = "EMAIL"
	NotifierTypeTelegram NotifierType = "TELEGRAM"
	NotifierTypeWebhook  NotifierType = "WEBHOOK"
	NotifierTypeSlack    NotifierType = "SLACK"
	NotifierTypeDiscord  NotifierType = "DISCORD"
	NotifierTypeTeams    NotifierType = "TEAMS"
)
