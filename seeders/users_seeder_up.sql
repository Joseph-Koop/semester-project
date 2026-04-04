INSERT INTO users (role, username, email, password_hash, activated) VALUES
    ('admin', 'Joseph Koop', 'jk@example.com', '\x24326124313224634d557143562e2f5332384643667547714d614c596535463945685137304e686c42482e6f77727166643633485544663778345869', true),
    ('trainer', 'Bethany Hill', 'bh@example.com', '\x24326124313224634d557143562e2f5332384643667547714d614c596535463945685137304e686c42482e6f77727166643633485544663778345869', true),
    ('member', 'Laura Goodwill', 'lg@example.com', '\x24326124313224634d557143562e2f5332384643667547714d614c596535463945685137304e686c42482e6f77727166643633485544663778345869', true),
    ('member', 'Paul Bunyan', 'pbexample.com', '\x24326124313224634d557143562e2f5332384643667547714d614c596535463945685137304e686c42482e6f77727166643633485544663778345869', false);

INSERT INTO permissions (code) VALUES 
    ('gyms:read'),
    ('gyms:write');
    ('studios:read'),
    ('studios:write');
    ('trainers:read'),
    ('trainers:write');
    ('classes:read'),
    ('classes:write');
    ('session_times:read'),
    ('session_times:write');
    ('members:read'),
    ('members:write');
    ('registrations:read'),
    ('registrations:write');
    ('sessions:read'),
    ('sessions:write');
    ('attendance:read'),
    ('attendance:write');
    ('users:read'),
    ('users:write');
    ('tokens:read'),
    ('tokens:write');
    ('permissions:read'),
    ('permissions:write');
    ('users_permissions:read'),
    ('users_permissions:write');

INSERT INTO users_permissions (user_id, permission_id) VALUES
    (1, 1),
    (1, 2),
    (2, 1),
    (3, 1);