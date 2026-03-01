create table if not exists passports
(
    id            serial primary key,
    number        varchar(64) unique       not null, -- TODO: check format
    issue_date    timestamp with time zone not null,
    name          varchar(128)             not null,
    second_name   varchar(128),
    surname       varchar(128)             not null,
    gender        varchar(16)              not null,
    birthday      timestamp                not null,
    birth_city    varchar(128)             not null,
    birth_country varchar(128)             not null
);

create table if not exists cards
(
    id          serial primary key,
    number      varchar(64) unique       not null, -- TODO: check format (all digits)
    expire_date timestamp with time zone not null,
    pin         smallint                 not null, -- TODO: check format
    name        varchar(128)             not null,
    surname     varchar(128)             not null
);

create table if not exists users
(
    id          serial primary key,
    passport_id int references passports (id),
    card_id     int references cards (id),
    email       varchar(256) unique,  -- TODO: check format
    phone       varchar(32) unique,-- TODO: check format
    password    varchar(128) not null -- TODO: check format (at least 8 characters)
);