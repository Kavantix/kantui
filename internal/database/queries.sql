-- name: GetTicketById :one
SELECT * FROM tickets
WHERE id = @id LIMIT 1;

-- name: GetTickets :many
SELECT * FROM tickets
order by rank, id;

-- name: AddTicket :one
insert into tickets (
  title, description, rank
)
values (
  @title, @description, (select coalesce(max(rank) + 1000000, 0) from tickets)
)
returning id, rank;

-- name: UpdateStatus :exec
update tickets
set status = @status
where id = @id;

-- name: UpdateRank :exec
update tickets
set rank = @rank
where id = @id;

-- name: UpdateTicketContent :exec
update tickets
set
  title = @title,
  description = @description
where id = @id;

-- name: DeleteTicket :exec
delete from tickets
where id = @id;
