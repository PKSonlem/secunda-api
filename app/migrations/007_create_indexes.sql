-- tasks: основные фильтры
CREATE INDEX idx_tasks_team_id     ON tasks (team_id);
CREATE INDEX idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX idx_tasks_status      ON tasks (status);
CREATE INDEX idx_tasks_created_by  ON tasks (created_by);
CREATE INDEX idx_tasks_updated_at  ON tasks (updated_at);

-- task_history: поиск по задаче
CREATE INDEX idx_th_task_id ON task_history (task_id);

-- task_comments: поиск по задаче
CREATE INDEX idx_tc_task_id ON task_comments (task_id);

-- team_members: поиск по пользователю
CREATE INDEX idx_tm_user_id ON team_members (user_id);
