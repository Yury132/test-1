-- +goose Up
create table if not exists public.service_user
(
    id         bigint       not null primary key,
    name       varchar(100) not null,
    email        varchar(100) not null
);

-- +goose Down
DROP TABLE public.service_user;