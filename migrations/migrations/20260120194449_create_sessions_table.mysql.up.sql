CREATE TABLE sessions (
	token CHAR(43) COLLATE utf8mb4_bin PRIMARY KEY,
	data BLOB NOT NULL,
	expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);