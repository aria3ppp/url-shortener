BEGIN;

CREATE TABLE IF NOT EXISTS links (
    shortened_string VARCHAR(40) PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    username VARCHAR(40) NOT NULL
);

-- add username foreign key constraint
ALTER TABLE IF EXISTS links
    ADD CONSTRAINT links_fk_users
    FOREIGN KEY (username)
    REFERENCES users(username);

COMMIT;