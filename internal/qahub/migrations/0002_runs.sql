CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> '')
);

CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> '')
);

CREATE INDEX executions_run_id ON executions (run_id);
