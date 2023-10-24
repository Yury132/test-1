-- +goose Up
CREATE TABLE IF NOT EXISTS public.service_user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT,
    name TEXT,
    surname TEXT
);

-- +goose Down
DROP TABLE public.service_user;