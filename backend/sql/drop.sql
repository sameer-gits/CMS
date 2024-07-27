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

-- index
DROP INDEX IF EXISTS idx_username_users;

DROP INDEX IF EXISTS idx_name_categories;

DROP INDEX IF EXISTS idx_title_articles;

DROP INDEX IF EXISTS idx_author_id_articles;

DROP INDEX IF EXISTS idx_author_id_messages;

DROP INDEX IF EXISTS idx_in_table_meesages;

DROP INDEX IF EXISTS idx_in_table_id_messages;

DROP INDEX IF EXISTS idx_forum_name_forums;

DROP INDEX IF EXISTS idx_forum_id_users_forums;

DROP INDEX IF EXISTS idx_forum_id_admins_forums;

DROP INDEX IF EXISTS idx_forum_id_mods_forums;

DROP INDEX IF EXISTS idx_poll_id_polls;

DROP INDEX IF EXISTS idx_poll_id_options;

DROP INDEX IF EXISTS idx_poll_id_votes;