ALTER TABLE cases ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE cases ADD COLUMN priority TEXT NOT NULL DEFAULT 'Medium' CHECK (priority IN ('Low', 'Medium', 'High'));
ALTER TABLE executions ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE executions ADD COLUMN priority TEXT NOT NULL DEFAULT 'Medium' CHECK (priority IN ('Low', 'Medium', 'High'));
