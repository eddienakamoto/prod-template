
-- +migrate Up
create table ttt (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- +migrate Down
drop table ttt;