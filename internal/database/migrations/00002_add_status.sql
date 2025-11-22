-- +goose Up
-- +goose StatementBegin
alter table tickets
  add column status text not null default 'TODO';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table tickets
  drop column status;
-- +goose StatementEnd

