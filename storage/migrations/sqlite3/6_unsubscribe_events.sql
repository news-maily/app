-- +migrate Up

CREATE TABLE IF NOT EXISTS `unsubscribe_events`
(
    `id`         varbinary primary key,
    `email`      varchar(191) NOT NULL,
    `created_at` datetime     NOT NULL
);

-- +migrate Down

DROP TABLE `unsubscribe_events`;