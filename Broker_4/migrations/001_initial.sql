CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    description TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS tasks_project_id ON tasks (project_id);
CREATE TABLE IF NOT EXISTS processed_requests (
    scope TEXT NOT NULL,
    id TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    response BLOB NOT NULL,
    PRIMARY KEY (scope, id)
);
