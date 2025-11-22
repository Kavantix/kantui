-- +goose Up
-- +goose StatementBegin
CREATE TABLE tickets (
  id   INTEGER PRIMARY KEY,
  title text    NOT NULL,
  description  text
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tickets;
-- +goose StatementEnd

