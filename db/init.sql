drop table if exists passports;
create table if not exists passports (
    id serial primary key,
    name va
);

create table if not exists users (
    id serial primary key,
    foreign key (passport_id) references passports(id)
);
