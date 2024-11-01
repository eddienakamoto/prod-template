
-- +migrate Up
CREATE TABLE another (
    id SERIAL PRIMARY KEY
);

-- +migrate Down
DROP TABLE another;