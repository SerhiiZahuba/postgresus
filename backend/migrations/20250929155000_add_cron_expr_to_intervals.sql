-- +goose Up
-- +goose StatementBegin
ALTER TABLE intervals
    ADD COLUMN cron_expr text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE intervals
DROP COLUMN IF EXISTS cron_expr;
-- +goose StatementEnd