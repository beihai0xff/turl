-- auto-generated definition
create table sequences
(
    id         bigint unsigned auto_increment
        primary key,
    created_at datetime(3)  null,
    updated_at datetime(3)  null,
    deleted_at datetime(3)  null,
    name       varchar(500) not null,
    sequence   bigint       not null,
    version    bigint       null,
    constraint idx_sequences_name
        unique (name)
);

create index idx_sequences_deleted_at
    on sequences (deleted_at);

