-- user table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    user_identifier UUID DEFAULT gen_random_uuid () UNIQUE,
    username VARCHAR(64) NOT NULL UNIQUE,
    fullname VARCHAR(64) NOT NULL,
    role CHAR CHECK (role IN ('A', 'S')) NOT NULL DEFAULT 'S',
    email VARCHAR(128) NOT NULL UNIQUE,
    profile_image BYTEA,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    password_hash VARCHAR(96) NOT NULL
);

-- article category table
CREATE TABLE IF NOT EXISTS categories (
    category_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    category_name VARCHAR(64) NOT NULL UNIQUE
);

-- article table
CREATE TABLE IF NOT EXISTS articles (
    author_identifier UUID NOT NULL,
    author VARCHAR(64) NOT NULL,
    category_id UUID,
    title VARCHAR(256) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    article_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    FOREIGN KEY (category_id) REFERENCES categories (category_id) ON DELETE SET NULL
);

-- message or comment
CREATE TABLE IF NOT EXISTS messages (
    author VARCHAR(64) NOT NULL,
    author_identifier UUID NOT NULL,
    message_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    reply_to_identifier UUID,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    in_table CHAR CHECK (in_table IN ('F', 'A', 'P')) NOT NULL,
    in_table_id UUID NOT NULL
);

-- forum table
CREATE TABLE IF NOT EXISTS forums (
    forum_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    forum_name VARCHAR(128) NOT NULL UNIQUE,
    forum_image BYTEA,
    public BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by_identifier UUID NOT NULL
);

-- forum_users | many to many
CREATE TABLE IF NOT EXISTS forum_users (
    user_identifier UUID REFERENCES users (user_identifier) ON DELETE CASCADE,
    forum_id UUID REFERENCES forums (forum_id) ON DELETE CASCADE,
    PRIMARY KEY (user_identifier, forum_id)
);

-- forum_admins | many to many
CREATE TABLE IF NOT EXISTS forum_admins (
    user_identifier UUID REFERENCES users (user_identifier) ON DELETE CASCADE,
    forum_id UUID REFERENCES forums (forum_id) ON DELETE CASCADE,
    PRIMARY KEY (user_identifier, forum_id)
);

-- forum_mods | many to many
CREATE TABLE IF NOT EXISTS forum_mods (
    user_identifier UUID REFERENCES users (user_identifier) ON DELETE CASCADE,
    forum_id UUID REFERENCES forums (forum_id) ON DELETE CASCADE,
    PRIMARY KEY (user_identifier, forum_id)
);

-- poll table
CREATE TABLE IF NOT EXISTS polls (
    poll_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    poll_title VARCHAR(256) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by_identifier UUID NOT NULL
);

-- poll options
CREATE TABLE IF NOT EXISTS poll_options (
    option_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    poll_id UUID REFERENCES polls (poll_id) ON DELETE CASCADE,
    option_text VARCHAR(256) NOT NULL,
    votes INTEGER DEFAULT 0
);

-- poll votes
CREATE TABLE IF NOT EXISTS poll_votes (
    vote_id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
    poll_id UUID REFERENCES polls (poll_id) ON DELETE CASCADE,
    option_id UUID REFERENCES poll_options (option_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    voter_identifier UUID NOT NULL,
    UNIQUE (poll_id, voter_identifier)
);

-- index
CREATE INDEX IF NOT EXISTS idx_username_users ON users (username);

CREATE INDEX IF NOT EXISTS idx_user_identifier_users ON users (user_identifier);

CREATE INDEX IF NOT EXISTS idx_name_categories ON categories (category_name);

CREATE INDEX IF NOT EXISTS idx_title_articles ON articles (title);

CREATE INDEX IF NOT EXISTS idx_author_identifier_articles ON articles (author_identifier);

CREATE INDEX IF NOT EXISTS idx_author_identifier_messages ON messages (author_identifier);

CREATE INDEX IF NOT EXISTS idx_reply_to_identifier_messages ON messages (reply_to_identifier);

CREATE INDEX IF NOT EXISTS idx_in_table_meesages ON messages (in_table);

CREATE INDEX IF NOT EXISTS idx_in_table_id_messages ON messages (in_table_id);

CREATE INDEX IF NOT EXISTS idx_forum_name_forums ON forums (forum_name);

CREATE INDEX IF NOT EXISTS idx_forum_id_users_forums ON forum_users (forum_id);

CREATE INDEX IF NOT EXISTS idx_forum_id_admins_forums ON forum_admins (forum_id);

CREATE INDEX IF NOT EXISTS idx_forum_id_mods_forums ON forum_mods (forum_id);

CREATE INDEX IF NOT EXISTS idx_poll_id_polls ON polls (poll_id);

CREATE INDEX IF NOT EXISTS idx_poll_id_options ON poll_options (poll_id);

CREATE INDEX IF NOT EXISTS idx_poll_id_votes ON poll_votes (poll_id);