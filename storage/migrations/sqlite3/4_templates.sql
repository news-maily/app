-- +migrate Up

CREATE TABLE IF NOT EXISTS "templates"
(
    "id"         integer primary key autoincrement,
    `user_id`    integer unsigned NOT NULL,
    `name`       varchar(191)     NOT NULL,
    "subject_part"    varchar(191)     NOT NULL,
    `text_part`  text,
    "created_at" datetime,
    "updated_at" datetime
);

-- +migrate Down

DROP TABLE "templates";
