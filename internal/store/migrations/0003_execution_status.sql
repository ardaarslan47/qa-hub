ALTER TABLE executions ADD COLUMN status TEXT CHECK (status IN ('Passed', 'Failed'));
