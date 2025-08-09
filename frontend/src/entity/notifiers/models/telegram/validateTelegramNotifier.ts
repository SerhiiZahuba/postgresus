import type { TelegramNotifier } from './TelegramNotifier';

export const validateTelegramNotifier = (notifier: TelegramNotifier): boolean => {
  if (!notifier.botToken) {
    return false;
  }

  if (!notifier.targetChatId) {
    return false;
  }

  // If thread is enabled, thread ID must be present and valid
  if (notifier.isSendToThreadEnabled && (!notifier.threadId || notifier.threadId <= 0)) {
    return false;
  }

  return true;
};
