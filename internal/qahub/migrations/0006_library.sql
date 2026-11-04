DROP INDEX executions_run_id;
ALTER TABLE executions RENAME TO old_executions;
ALTER TABLE cases RENAME TO old_cases;

CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK (title <> '')
);

CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed'))
);

INSERT INTO cases (id, title)
SELECT id, title FROM old_cases;

INSERT INTO executions (id, run_id, case_id, title, status)
SELECT id, run_id, case_id, title, status FROM old_executions;

DROP TABLE old_executions;
DROP TABLE old_cases;

CREATE INDEX executions_run_id ON executions (run_id);
