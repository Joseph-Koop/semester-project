CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);
