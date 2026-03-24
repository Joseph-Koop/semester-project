
CREATE TYPE access_type AS ENUM ('general', 'classes');
CREATE TYPE membership_tier_type AS ENUM ('basic', 'standard', 'premium');
CREATE TYPE registration_status_type AS ENUM ('active', 'dropped');
CREATE TYPE day_type AS ENUM ('sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat');


CREATE TABLE IF NOT EXISTS gyms (
    id SERIAL PRIMARY KEY,
    location VARCHAR(255) NOT NULL,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS studios (
    id SERIAL PRIMARY KEY,
    gym_id INTEGER NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    access access_type NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS trainers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    address TEXT,
    phone VARCHAR(30),
    email VARCHAR(150) UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    studio_id INTEGER REFERENCES studios(id) ON DELETE SET NULL,
    trainer_id INTEGER NOT NULL REFERENCES trainers(id) ON DELETE SET NULL,
    capacity_limit INTEGER NOT NULL CHECK (capacity_limit > 0),
    membership_tier membership_tier_type NOT NULL,
    name VARCHAR(100) NOT NULL,
    terminated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS session_times (
    id SERIAL PRIMARY KEY,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    day day_type NOT NULL,
    time TIME NOT NULL,
    duration INTEGER NOT NULL,                  --minutes
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    address TEXT,
    phone VARCHAR(30),
    email VARCHAR(150) UNIQUE,
    membership_tier membership_tier_type NOT NULL,
    expiry_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS registrations (
    id SERIAL PRIMARY KEY,
    member_id INTEGER NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    status registration_status_type NOT NULL DEFAULT 'active',
    UNIQUE (member_id, class_id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS attendance (
    id SERIAL PRIMARY KEY,
    registration_id INTEGER NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    UNIQUE (registration_id, session_id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);
