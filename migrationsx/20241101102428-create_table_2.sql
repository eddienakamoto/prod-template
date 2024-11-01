
-- +migrate Up
create table again (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- +migrate Down
drop table again;