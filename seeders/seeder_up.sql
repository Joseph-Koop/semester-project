INSERT INTO gyms (location, name) VALUES
('Downtown', 'Iron Temple Gym'),
('Uptown', 'Pulse Fitness Center');

INSERT INTO studios (gym_id, name, access) VALUES
(1, 'Main Floor', 'general'),
(1, 'Yoga Studio', 'classes'),
(2, 'Cardio Room', 'general'),
(2, 'Spin Studio', 'classes');

INSERT INTO trainers (name, address, phone, email) VALUES
('Michael Grant', '123 Fitness St', '5015551001', 'michael@iron.com'),
('Sarah Lopez', '45 Wellness Ave', '5015551002', 'sarah@iron.com'),
('Daniel Reed', '99 Strength Rd', '5015552001', 'daniel@pulse.com'),
('Emily Cruz', '12 Cardio Blvd', '5015552002', 'emily@pulse.com');

INSERT INTO classes (studio_id, trainer_id, capacity_limit, membership_tier, name, terminated) VALUES
(2, 2, 20, 'basic', 'Morning Yoga', FALSE),
(4, 4, 15, 'standard', 'Spin Blast', FALSE),
(2, 1, 25, 'premium', 'Advanced Pilates', FALSE),
(4, 3, 30, 'standard', 'HIIT Power', FALSE),
(1, 1, 12, 'basic', 'Sunrise Stretch', FALSE),
(3, 2, 40, 'premium', 'Strength Lab', FALSE),
(2, 4, 18, 'standard', 'Core Crusher', FALSE),
(1, 3, 50, 'premium', 'Elite Bootcamp', FALSE),
(3, 1, 22, 'basic', 'Beginner Boxing', FALSE),
(4, 2, 35, 'standard', 'Cardio Burn', FALSE),
(2, 3, 60, 'premium', 'Powerlifting Intro', FALSE),
(1, 4, 28, 'basic', 'Mobility Flow', TRUE),
(3, 4, 75, 'premium', 'Athlete Conditioning', FALSE),
(4, 1, 45, 'standard', 'Total Body Blast', FALSE),
(1, 2, 16, 'basic', 'Evening Relax Yoga', TRUE),
(2, 2, 90, 'premium', 'Max Strength', FALSE),
(3, 3, 55, 'standard', 'Functional Fitness', FALSE),
(4, 4, 70, 'premium', 'Performance Lab', TRUE),
(1, 3, 10, 'basic', 'Gentle Stretch', FALSE),
(2, 1, 33, 'standard', 'Circuit Express', FALSE);

INSERT INTO session_times (class_id, day, time) VALUES
(1, 'mon', '08:00'),
(1, 'wed', '08:00'),
(2, 'tue', '18:00'),
(2, 'thu', '18:00'),
(3, 'fri', '17:00'),
(4, 'sat', '09:00'),
(4, 'sun', '09:00');

INSERT INTO members (name, address, phone, email, membership_tier, expiry_date) VALUES
('John Carter', '12 Palm St', '5015553001', 'john@gmail.com', 'basic', '2026-12-31'),
('Lisa Morgan', '34 Pine Ave', '5015553002', 'lisa@gmail.com', 'standard', '2026-10-15'),
('Robert King', '78 Oak Drive', '5015553003', 'robert@gmail.com', 'premium', '2027-01-01'),
('Natalie Green', '22 Cedar Rd', '5015553004', 'natalie@gmail.com', 'standard', '2026-08-20');

INSERT INTO registrations (member_id, class_id, status) VALUES
(1, 1, 'active'),
(2, 2, 'active'),
(3, 3, 'active'),
(4, 4, 'active'),
(2, 4, 'dropped');

INSERT INTO sessions (class_id) VALUES
(1),
(2),
(3),
(4),
(1),
(2),
(3),
(4),
(1),
(2),
(3),
(4),
(1),
(2),
(3),
(4);

INSERT INTO attendance (registration_id, session_id) VALUES
(1, 1),
(1, 2),
(2, 3),
(2, 4),
(3, 5),
(4, 6),
(4, 7);