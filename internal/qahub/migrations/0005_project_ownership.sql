INSERT OR IGNORE INTO projects (name)
SELECT 'Default'
WHERE EXISTS (SELECT 1 FROM cases) OR EXISTS (SELECT 1 FROM runs);

DROP INDEX executions_run_id;
ALTER TABLE executions RENAME TO old_executions;
ALTER TABLE runs RENAME TO old_runs;
ALTER TABLE cases RENAME TO old_cases;

CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    title TEXT NOT NULL CHECK (title <> '')
);

CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    name TEXT NOT NULL CHECK (name <> '')
);

CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed'))
);

INSERT INTO cases (id, project_id, title)
SELECT id, (SELECT id FROM projects WHERE name = 'Default'), title FROM old_cases;

INSERT INTO runs (id, project_id, name)
SELECT id, (SELECT id FROM projects WHERE name = 'Default'), name FROM old_runs;

INSERT INTO executions (id, run_id, case_id, title, status)
SELECT id, run_id, case_id, title, status FROM old_executions;

DROP TABLE old_executions;
DROP TABLE old_runs;
DROP TABLE old_cases;

CREATE INDEX cases_project_id ON cases (project_id);
CREATE INDEX runs_project_id ON runs (project_id);
CREATE INDEX executions_run_id ON executions (run_id);
