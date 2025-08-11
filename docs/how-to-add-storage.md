# How to add new storage to Postgresus (S3, FTP, Google Drive, NAS, etc.)

## Backend part

1. Create new model in `backend/internal/features/storages/models/{storage_name}/` folder. Implement `StorageFileSaver` interface from parent folder.
   - The model should implement `SaveFile(logger *slog.Logger, fileID uuid.UUID, file io.Reader) error`, `GetFile(fileID uuid.UUID) (io.ReadCloser, error)`, `DeleteFile(fileID uuid.UUID) error`, `Validate() error`, and `TestConnection() error` methods
   - Use UUID primary key as `StorageID` that references the main storages table
   - Add `TableName() string` method to return the proper table name

2. Add new storage type to `backend/internal/features/storages/enums.go` in the `StorageType` constants.

3. Update the main `Storage` model in `backend/internal/features/storages/model.go`:
   - Add new storage field with GORM foreign key relation
   - Update `getSpecificStorage()` method to handle the new type
   - Update `SaveFile()`, `GetFile()`, and `DeleteFile()` methods to route to the new storage
   - Update `Validate()` method to include new storage validation

4. If you need to add some .env variables to test, add them in `backend/internal/config/config.go` (so we can use it in tests)

5. If you need some Docker container to test, add it to `backend/docker-compose.yml.example`. For sensitive data - keep it blank.

6. If you need some sensitive envs to test in pipeline, message @rostislav_dugin so I can add it to GitHub Actions. For example, Google Drive envs or FTP credentials.

7. Create new migration in `backend/migrations` folder:
   - Create table with `storage_id` as UUID primary key
   - Add foreign key constraint to `storages` table with CASCADE DELETE
   - Look at existing storage migrations for reference

8. Update tests in `backend/internal/features/storages/model_test.go` to test new storage

9. Make sure that all tests are passing.

## Frontend part

If you are able to develop only backend - it's fine, message @rostislav_dugin so I can complete UI part.

1. Add models and api to `frontend/src/entity/storages/models/` folder and update `index.ts` file to include new model exports.
   - Create TypeScript interface for your storage model
   - Add validation function if needed

2. Upload an SVG icon to `public/icons/storages/`, update `src/entity/storages/models/getStorageLogoFromType.ts` to return new icon path, update `src/entity/storages/models/StorageType.ts` to include new type, and update `src/entity/storages/models/getStorageNameFromType.ts` to return new name.

3. Add UI components to manage your storage:
   - `src/features/storages/ui/edit/storages/Edit{StorageName}Component.tsx` (for editing)
   - `src/features/storages/ui/show/storages/Show{StorageName}Component.tsx` (for display)

4. Update main components to handle the new storage type:
   - `EditStorageComponent.tsx` - add import and component rendering
   - `ShowStorageComponent.tsx` - add import and component rendering

5. Make sure everything is working as expected.
