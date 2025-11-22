-- +goose Up
-- +goose StatementBegin
alter table tickets
  add column rank integer not null default 0;

update tickets
  set rank = id * 1000000;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table tickets
  drop column rank;
-- +goose StatementEnd

