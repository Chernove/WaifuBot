CREATE TABLE users (
    "id" SERIAL NOT NULL,
    "user_id" BIGINT NOT NULL PRIMARY KEY,
    "quote" TEXT NOT NULL DEFAULT '',
    "date" TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:00-00',
    "favorite" BIGINT -- "tokens" INT NOT NULL DEFAULT 0
);
CREATE TABLE characters (
    "user_id" BIGINT NOT NULL,
    "id" BIGINT NOT NULL,
    "image" CHARACTER VARYING(256) NOT NULL DEFAULT '',
    "name" CHARACTER VARYING(128) NOT NULL DEFAULT '',
    "date" TIMESTAMP NOT NULL DEFAULT NOW(),
    "type" VARCHAR NOT NULL DEFAULT '',
    PRIMARY KEY ("id", "user_id"),
    CONSTRAINT "users_characters_fk" FOREIGN KEY (user_id) REFERENCES users (user_id)
);
ALTER TABLE users
ADD CONSTRAINT "characters_users_fk" FOREIGN KEY (favorite, user_id) REFERENCES characters (id, user_id);