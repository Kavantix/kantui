-- +goose Up
-- +goose StatementBegin
CREATE TABLE tickets (
  id          integer primary key autoincrement,
  status      text not null default 'TODO',
  title       text not null,
  description text
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tickets;
-- +goose StatementEnd

