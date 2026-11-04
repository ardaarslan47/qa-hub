CREATE TABLE case_preconditions (
    case_id INTEGER NOT NULL REFERENCES cases (id),
    precondition_id INTEGER NOT NULL REFERENCES cases (id),
    PRIMARY KEY (case_id, precondition_id)
);

CREATE TABLE execution_preconditions (
    id INTEGER PRIMARY KEY,
    execution_id INTEGER NOT NULL REFERENCES executions (id),
    title TEXT NOT NULL CHECK (title <> '')
);

CREATE INDEX execution_preconditions_execution_id ON execution_preconditions (execution_id);
