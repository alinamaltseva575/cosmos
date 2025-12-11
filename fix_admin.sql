-- fix_admin.sql
TRUNCATE TABLE users RESTART IDENTITY;

INSERT INTO users (username, email, password_hash, role, created_at)
VALUES (
    'admin',
    'admin@cosmos.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- пароль: password
    'admin',
    NOW()
);
