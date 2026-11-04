ALTER TABLE execution_preconditions RENAME TO old_execution_preconditions;
ALTER TABLE executions RENAME TO old_executions;
ALTER TABLE case_preconditions RENAME TO old_case_preconditions;
ALTER TABLE cases RENAME TO old_cases;

CREATE TABLE cases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL CHECK (title <> ''),
    description TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'Medium' CHECK (priority IN ('Low', 'Medium', 'High'))
);

CREATE TABLE case_preconditions (
    case_id INTEGER NOT NULL REFERENCES cases (id) ON DELETE CASCADE,
    precondition_id INTEGER NOT NULL REFERENCES cases (id) ON DELETE CASCADE,
    PRIMARY KEY (case_id, precondition_id)
);

CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER REFERENCES cases (id) ON DELETE SET NULL,
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed')),
    description TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'Medium' CHECK (priority IN ('Low', 'Medium', 'High'))
);

CREATE TABLE execution_preconditions (
    id INTEGER PRIMARY KEY,
    execution_id INTEGER NOT NULL REFERENCES executions (id),
    title TEXT NOT NULL CHECK (title <> '')
);

INSERT INTO cases (id, title, description, priority)
SELECT id, title, description, priority FROM old_cases;

INSERT INTO case_preconditions (case_id, precondition_id)
SELECT case_id, precondition_id FROM old_case_preconditions;

INSERT INTO executions (id, run_id, case_id, title, status, description, priority)
SELECT id, run_id, case_id, title, status, description, priority FROM old_executions;

INSERT INTO execution_preconditions (id, execution_id, title)
SELECT id, execution_id, title FROM old_execution_preconditions;

DROP TABLE old_execution_preconditions;
DROP TABLE old_executions;
DROP TABLE old_case_preconditions;
DROP TABLE old_cases;

CREATE INDEX executions_run_id ON executions (run_id);
CREATE INDEX execution_preconditions_execution_id ON execution_preconditions (execution_id);
