-- +goose Up
CREATE TABLE sites
(
  ID SERIAL PRIMARY KEY,
  IP text NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS sites;
