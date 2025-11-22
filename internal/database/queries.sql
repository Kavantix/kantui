-- name: GetTicketById :one
SELECT * FROM tickets
WHERE id = @id LIMIT 1;

-- name: GetTickets :many
SELECT * FROM tickets;

-- name: AddTicket :one
insert into tickets (
  title, description
)
values (
  @title, @description
)
returning id;

-- name: UpdateStatus :exec
update tickets
set status = @status
where id = @id;
