CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK (title <> '')
);
