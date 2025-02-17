CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username varchar(255) NOT NULL,
    email citext,
    password bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now()
);
