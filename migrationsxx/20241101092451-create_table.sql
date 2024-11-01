
-- +migrate Up
CREATE TABLE dummy (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- +migrate Down
DROP TABLE dummy;