package main

import (
	"context"
	"fmt"
)

func createSchema() error {
	schema := `
    DROP TABLE IF EXISTS articles;
    DROP TABLE IF EXISTS categories;
    DROP TABLE IF EXISTS messages;
    DROP TABLE IF EXISTS poll_votes;
    DROP TABLE IF EXISTS poll_options;
    DROP TABLE IF EXISTS polls;
    DROP TABLE IF EXISTS forum_users;
    DROP TABLE IF EXISTS forum_admins;
    DROP TABLE IF EXISTS forum_mods;
    DROP TABLE IF EXISTS forums;
    DROP TABLE IF EXISTS users;

    -- user table
    CREATE TABLE users (
        user_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        username VARCHAR(64) NOT NULL UNIQUE,
        fullname VARCHAR(64) NOT NULL,
        role CHAR CHECK (role IN ('A', 'S')) NOT NULL DEFAULT 'S',
        email VARCHAR(128) NOT NULL UNIQUE,
        profile_image BYTEA,
        joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        password_hash BYTEA NOT NULL
    );

    -- article category table
    CREATE TABLE categories (
        category_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        name VARCHAR(64) NOT NULL UNIQUE
    );

    -- article table
    CREATE TABLE articles (
        author VARCHAR(64) NOT NULL,
        category_id UUID,
        title VARCHAR(256) NOT NULL UNIQUE,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        article_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE SET NULL
    );

    -- message or comment
    CREATE TABLE messages (
        author VARCHAR(64) NOT NULL,
        message_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        reply_to UUID,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        in_table CHAR CHECK (in_table IN ('F', 'A', 'P')) NOT NULL,
        FOREIGN KEY (reply_to) REFERENCES messages(message_id) ON DELETE CASCADE
    );

    -- forum table
    CREATE TABLE forums (
        forum_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        forum_name VARCHAR(128) NOT NULL UNIQUE,
        forum_image BYTEA NOT NULL,
        public BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_by VARCHAR(64) NOT NULL
        );

    -- forum_users | many to many
    CREATE TABLE forum_users (
        user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
        forum_id UUID REFERENCES forums(forum_id) ON DELETE CASCADE,
        PRIMARY KEY (user_id, forum_id)
    );

    -- forum_admins | many to many
    CREATE TABLE forum_admins (
        user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
        forum_id UUID REFERENCES forums(forum_id) ON DELETE CASCADE,
        PRIMARY KEY (user_id, forum_id)
    );

    -- forum_mods | many to many
    CREATE TABLE forum_mods (
        user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
        forum_id UUID REFERENCES forums(forum_id) ON DELETE CASCADE,
        PRIMARY KEY (user_id, forum_id)
    );

    -- poll table
    CREATE TABLE polls (
        poll_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        poll_title VARCHAR(256) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_by VARCHAR(64) NOT NULL
    );

    -- poll options
    CREATE TABLE poll_options (
        option_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        poll_id UUID REFERENCES polls(poll_id) ON DELETE CASCADE,
        option_text VARCHAR(256) NOT NULL,
        votes INTEGER DEFAULT 0
    );

    -- poll votes
    CREATE TABLE poll_votes (
        vote_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
        poll_id UUID REFERENCES polls(poll_id) ON DELETE CASCADE,
        option_id UUID REFERENCES poll_options(option_id) ON DELETE CASCADE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        voter VARCHAR(64) NOT NULL,
        UNIQUE (poll_id, voter)
    );
    
    CREATE INDEX idx_username ON users (username);
    CREATE INDEX idx_name ON categories (name);
    CREATE INDEX idx_title ON articles (title);
    CREATE INDEX idx_author_article ON articles (author);
    CREATE INDEX idx_in_table ON messages (in_table);
    CREATE INDEX idx_author_messages ON messages (author);
    CREATE INDEX idx_forum_name ON forums (forum_name);
    CREATE INDEX idx_forum_id_users ON forum_users (forum_id);
    CREATE INDEX idx_forum_id_admins ON forum_admins (forum_id);
    CREATE INDEX idx_forum_id_mods ON forum_mods (forum_id);
    CREATE INDEX idx_poll_id_polls ON polls (poll_id);
    CREATE INDEX idx_poll_id_options ON poll_options (poll_id);
    CREATE INDEX idx_poll_id_votes ON poll_votes (poll_id);
    `
	_, err := Pool.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("faild to make schema: %v", err)
	}

	return nil
}
