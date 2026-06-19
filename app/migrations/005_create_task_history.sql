CREATE TABLE IF NOT EXISTS task_history (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_id    BIGINT UNSIGNED NOT NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    field_name VARCHAR(50)     NOT NULL,
    old_value  TEXT,
    new_value  TEXT,
    changed_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_th_task FOREIGN KEY (task_id)    REFERENCES tasks (id) ON DELETE CASCADE,
    CONSTRAINT fk_th_user FOREIGN KEY (changed_by) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
