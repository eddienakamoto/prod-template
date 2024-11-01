
-- +migrate Up
create table againn (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- +migrate Down
drop table againn;