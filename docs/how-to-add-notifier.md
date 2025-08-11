# How to add new notifier to Postgresus (Discord, Slack, Telegram, Email, Webhook, etc.)

## Backend part

1. Create new model in `backend/internal/features/notifiers/models/{notifier_name}/` folder. Implement `NotificationSender` interface from parent folder.
   - The model should implement `Send(logger *slog.Logger, heading string, message string) error` and `Validate() error` methods
   - Use UUID primary key as `NotifierID` that references the main notifiers table

2. Add new notifier type to `backend/internal/features/notifiers/enums.go` in the `NotifierType` constants.

3. Update the main `Notifier` model in `backend/internal/features/notifiers/model.go`:
   - Add new notifier field with GORM foreign key relation
   - Update `getSpecificNotifier()` method to handle the new type
   - Update `Send()` method to route to the new notifier

4. If you need to add some .env variables to test, add them in `backend/internal/config/config.go` (so we can use it in tests)

5. If you need some Docker container to test, add it to `backend/docker-compose.yml.example`. For sensitive data - keep it blank.

6. If you need some sensitive envs to test in pipeline, message @rostislav_dugin so I can add it to GitHub Actions. For example, API keys or credentials.

7. Create new migration in `backend/migrations` folder:
   - Create table with `notifier_id` as UUID primary key
   - Add foreign key constraint to `notifiers` table with CASCADE DELETE
   - Look at existing notifier migrations for reference

8. Make sure that all tests are passing.

## Frontend part

If you are able to develop only backend - it's fine, message @rostislav_dugin so I can complete UI part.

1. Add models and validator to `frontend/src/entity/notifiers/models/{notifier_name}/` folder and update `index.ts` file to include new model exports.

2. Upload an SVG icon to `public/icons/notifiers/`, update `src/entity/notifiers/models/getNotifierLogoFromType.ts` to return new icon path, update `src/entity/notifiers/models/NotifierType.ts` to include new type, and update `src/entity/notifiers/models/getNotifierNameFromType.ts` to return new name.

3. Add UI components to manage your notifier:
   - `src/features/notifiers/ui/edit/notifiers/Edit{NotifierName}Component.tsx` (for editing)
   - `src/features/notifiers/ui/show/notifier/Show{NotifierName}Component.tsx` (for display)

4. Update main components to handle the new notifier type:
   - `EditNotifierComponent.tsx` - add import, validation function, and component rendering
   - `ShowNotifierComponent.tsx` - add import and component rendering

5. Make sure everything is working as expected.
