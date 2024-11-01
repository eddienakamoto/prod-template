
-- +migrate Up
CREATE TABLE anotherr (
    id SERIAL PRIMARY KEY
);

-- +migrate Down
DROP TABLE anotherr;