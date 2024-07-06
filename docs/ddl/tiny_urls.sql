-- auto-generated definition
create table tiny_urls
(
    id         bigint unsigned auto_increment
        primary key,
    created_at datetime(3)  null,
    updated_at datetime(3)  null,
    deleted_at datetime(3)  null,
    long_url   varchar(500) not null,
    short      bigint       not null,
    constraint idx_tiny_urls_long_url
        unique (long_url),
    constraint idx_tiny_urls_short
        unique (short)
);

create index idx_tiny_urls_deleted_at
    on tiny_urls (deleted_at);

