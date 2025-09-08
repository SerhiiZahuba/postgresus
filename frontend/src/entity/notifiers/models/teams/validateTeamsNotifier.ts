import type { TeamsNotifier } from './TeamsNotifier';

export const validateTeamsNotifier = (notifier: TeamsNotifier): boolean => {
    if (!notifier?.powerAutomateUrl) {
        return false;
    }

    try {
        const u = new URL(notifier.powerAutomateUrl);
        if (u.protocol !== 'http:' && u.protocol !== 'https:') return false;
    } catch {
        return false;
    }

    return true;
};
