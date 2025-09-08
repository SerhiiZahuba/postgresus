package notifiers

import (
	"postgresus-backend/internal/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotifierRepository struct{}

func (r *NotifierRepository) Save(notifier *Notifier) (*Notifier, error) {
	db := storage.GetDb()

	err := db.Transaction(func(tx *gorm.DB) error {

		switch notifier.NotifierType {
		case NotifierTypeTelegram:
			if notifier.TelegramNotifier != nil {
				notifier.TelegramNotifier.NotifierID = notifier.ID
			}
		case NotifierTypeEmail:
			if notifier.EmailNotifier != nil {
				notifier.EmailNotifier.NotifierID = notifier.ID
			}
		case NotifierTypeWebhook:
			if notifier.WebhookNotifier != nil {
				notifier.WebhookNotifier.NotifierID = notifier.ID
			}
		case NotifierTypeSlack:
			if notifier.SlackNotifier != nil {
				notifier.SlackNotifier.NotifierID = notifier.ID
			}
		case NotifierTypeDiscord:
			if notifier.DiscordNotifier != nil {
				notifier.DiscordNotifier.NotifierID = notifier.ID
			}
		case NotifierTypeTeams:
			if notifier.TeamsNotifier != nil {
				notifier.TeamsNotifier.NotifierID = notifier.ID
			}
		}

		if notifier.ID == uuid.Nil {
			if err := tx.
				Omit(
					"TelegramNotifier",
					"EmailNotifier",
					"WebhookNotifier",
					"SlackNotifier",
					"DiscordNotifier",
					"TeamsNotifier",
				).
				Create(notifier).Error; err != nil {
				return err
			}
		} else {
			if err := tx.
				Omit(
					"TelegramNotifier",
					"EmailNotifier",
					"WebhookNotifier",
					"SlackNotifier",
					"DiscordNotifier",
					"TeamsNotifier",
				).
				Save(notifier).Error; err != nil {
				return err
			}
		}

		switch notifier.NotifierType {
		case NotifierTypeTelegram:
			if notifier.TelegramNotifier != nil {
				notifier.TelegramNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.TelegramNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeEmail:
			if notifier.EmailNotifier != nil {
				notifier.EmailNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.EmailNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeWebhook:
			if notifier.WebhookNotifier != nil {
				notifier.WebhookNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.WebhookNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeSlack:
			if notifier.SlackNotifier != nil {
				notifier.SlackNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.SlackNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeDiscord:
			if notifier.DiscordNotifier != nil {
				notifier.DiscordNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.DiscordNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeTeams:
			if notifier.TeamsNotifier != nil {
				notifier.TeamsNotifier.NotifierID = notifier.ID
				if err := tx.Save(notifier.TeamsNotifier).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return notifier, nil
}

func (r *NotifierRepository) FindByID(id uuid.UUID) (*Notifier, error) {
	var notifier Notifier

	if err := storage.
		GetDb().
		Preload("TelegramNotifier").
		Preload("EmailNotifier").
		Preload("WebhookNotifier").
		Preload("SlackNotifier").
		Preload("DiscordNotifier").
		Preload("TeamsNotifier").
		Where("id = ?", id).
		First(&notifier).Error; err != nil {
		return nil, err
	}

	return &notifier, nil
}

func (r *NotifierRepository) FindByUserID(userID uuid.UUID) ([]*Notifier, error) {
	var notifiers []*Notifier

	if err := storage.
		GetDb().
		Preload("TelegramNotifier").
		Preload("EmailNotifier").
		Preload("WebhookNotifier").
		Preload("SlackNotifier").
		Preload("DiscordNotifier").
		Preload("TeamsNotifier").
		Where("user_id = ?", userID).
		Order("name ASC").
		Find(&notifiers).Error; err != nil {
		return nil, err
	}

	return notifiers, nil
}

func (r *NotifierRepository) Delete(notifier *Notifier) error {
	return storage.GetDb().Transaction(func(tx *gorm.DB) error {

		switch notifier.NotifierType {
		case NotifierTypeTelegram:
			if notifier.TelegramNotifier != nil {
				if err := tx.Delete(notifier.TelegramNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeEmail:
			if notifier.EmailNotifier != nil {
				if err := tx.Delete(notifier.EmailNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeWebhook:
			if notifier.WebhookNotifier != nil {
				if err := tx.Delete(notifier.WebhookNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeSlack:
			if notifier.SlackNotifier != nil {
				if err := tx.Delete(notifier.SlackNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeDiscord:
			if notifier.DiscordNotifier != nil {
				if err := tx.Delete(notifier.DiscordNotifier).Error; err != nil {
					return err
				}
			}
		case NotifierTypeTeams:
			if notifier.TeamsNotifier != nil {
				if err := tx.Delete(notifier.TeamsNotifier).Error; err != nil {
					return err
				}
			}
		}

		return tx.Delete(notifier).Error
	})
}
