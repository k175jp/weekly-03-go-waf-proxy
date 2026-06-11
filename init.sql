USE security_lab;

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL
);

INSERT INTO users (name, email) VALUES
('admin', 'admin@security.lab'),
('alice', 'alice@security.lab'),
('bob', 'bob@security.lab'),
('flag_user', 'FLAG{waf_bypass_success_2026}');