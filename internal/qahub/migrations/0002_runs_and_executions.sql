CREATE TABLE runs (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL CHECK (trim(name) <> '')
);

CREATE TABLE executions (
  id INTEGER PRIMARY KEY,
  run_id INTEGER NOT NULL REFERENCES runs(id),
  case_id INTEGER NOT NULL REFERENCES cases(id),
  title TEXT NOT NULL,
  status TEXT CHECK (status IN ('Passed', 'Failed'))
);
