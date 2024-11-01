
-- +migrate Up
create table another (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- +migrate Down
drop table another;