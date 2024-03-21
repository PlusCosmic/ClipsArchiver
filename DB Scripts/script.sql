create table legends
(
    id         int auto_increment
        primary key,
    name       varchar(30) null,
    card_image varchar(50) null
);

create table maps
(
    id         int auto_increment
        primary key,
    name       varchar(30)  null,
    card_image varchar(70)  null,
    als_name   varchar(100) null
);

create table tags
(
    id   int auto_increment
        primary key,
    name varchar(32) null
);

create table users
(
    id            int auto_increment
        primary key,
    name          varchar(32) null,
    apex_username varchar(32) null,
    apex_uid      varchar(32) null
);

create table clips
(
    id                  int auto_increment
        primary key,
    owner_id            int                     not null,
    filename            varchar(128)            not null,
    is_processed        tinyint(1) default 0    not null,
    created_at          timestamp               not null,
    duration            int                     null,
    map                 int                     null,
    game_mode           enum ('pubs', 'ranked') null,
    legend              int                     null,
    match_history_found tinyint(1) default 0    not null,
    constraint clips_legend_id_fk
        foreign key (legend) references legends (id),
    constraint clips_maps_id_fk
        foreign key (map) references maps (id),
    constraint clips_users_id_fk
        foreign key (owner_id) references users (id)
);

create table clips_queue
(
    id          int auto_increment
        primary key,
    clip_id     int                                                   not null,
    status      enum ('pending', 'queued', 'transcoding', 'finished') not null,
    started_at  datetime                                              null,
    finished_at datetime                                              null,
    constraint clips_queue_clips_id_fk
        foreign key (clip_id) references clips (id)
);

create table clips_tags
(
    clip_id int not null,
    tag_id  int not null,
    primary key (clip_id, tag_id),
    constraint clips_tags_clips_id_fk
        foreign key (clip_id) references clips (id),
    constraint clips_tags_tags_id_fk
        foreign key (tag_id) references tags (id)
);

create table match_history
(
    id         int auto_increment
        primary key,
    user_id    int                     null,
    game_start datetime                null,
    game_end   datetime                null,
    map        int                     null,
    legend     int                     null,
    game_mode  enum ('Pubs', 'Ranked') null,
    constraint match_history_legend_id_fk
        foreign key (legend) references legends (id),
    constraint match_history_maps_id_fk
        foreign key (map) references maps (id),
    constraint match_history_users_id_fk
        foreign key (user_id) references users (id)
);


