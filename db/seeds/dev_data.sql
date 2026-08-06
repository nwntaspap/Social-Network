-- Seed data for development
-- All users have password: password123
-- Idempotent: uses INSERT OR IGNORE for all inserts

BEGIN TRANSACTION;

------------------------------------------------------------
-- USERS (5)
------------------------------------------------------------
INSERT OR IGNORE INTO users (id, username, email, password_hash, first_name, last_name, avatar_url, date_of_birth, about_me, is_private)
VALUES
  ('550e8400-e29b-41d4-a716-446655440001', 'alice',   'alice@example.com',   '$2a$12$oA6qQvZF9zHXLjnsqWDvfe6dLmRxnw8uUI5N0zrA1ErAR3tyyUdB6', 'Alice',   'Smith',   'https://api.dicebear.com/7.x/avataaars/svg?seed=alice',   '1990-05-15', 'Admin and community leader', 0),
  ('550e8400-e29b-41d4-a716-446655440002', 'bob',     'bob@example.com',     '$2a$12$oA6qQvZF9zHXLjnsqWDvfe6dLmRxnw8uUI5N0zrA1ErAR3tyyUdB6', 'Bob',     'Johnson', 'https://api.dicebear.com/7.x/avataaars/svg?seed=bob',     '1992-08-22', 'Tech enthusiast and writer', 0),
  ('550e8400-e29b-41d4-a716-446655440003', 'charlie', 'charlie@example.com', '$2a$12$oA6qQvZF9zHXLjnsqWDvfe6dLmRxnw8uUI5N0zrA1ErAR3tyyUdB6', 'Charlie', 'Brown',   'https://api.dicebear.com/7.x/avataaars/svg?seed=charlie', '1988-12-01', 'Private account holder', 1),
  ('550e8400-e29b-41d4-a716-446655440004', 'diana',   'diana@example.com',   '$2a$12$oA6qQvZF9zHXLjnsqWDvfe6dLmRxnw8uUI5N0zrA1ErAR3tyyUdB6', 'Diana',   'Prince',  'https://api.dicebear.com/7.x/avataaars/svg?seed=diana',   '1995-03-10', 'Event organizer and traveler', 0),
  ('550e8400-e29b-41d4-a716-446655440005', 'eve',     'eve@example.com',     '$2a$12$oA6qQvZF9zHXLjnsqWDvfe6dLmRxnw8uUI5N0zrA1ErAR3tyyUdB6', 'Eve',     'Adams',   'https://api.dicebear.com/7.x/avataaars/svg?seed=eve',     '1993-07-28', 'New member excited to join!', 0);

------------------------------------------------------------
-- TOPICS (8) — id is INTEGER AUTOINCREMENT, not UUID
------------------------------------------------------------
INSERT OR IGNORE INTO topics (user_id, title, content, image_path, visibility, group_id)
VALUES
  ('550e8400-e29b-41d4-a716-446655440001', 'Welcome to the Forum!',        'This is the official welcome topic. Introduce yourself here!',                  NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440002', 'Go vs Rust: Performance',       'Let''s discuss the performance characteristics of Go and Rust.',                 NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440003', 'My Private Thoughts',           'This is a private topic just for me.',                                          NULL, 1, NULL),
  ('550e8400-e29b-41d4-a716-446655440001', 'Community Guidelines',           'Please read and follow these guidelines when posting.',                        NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440004', 'Upcoming Hackathon',            'Join us for a weekend hackathon! Sign up below.',                               NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440002', 'Best VS Code Extensions',       'Share your favorite VS Code extensions for productivity.',                      NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440005', 'New Member Questions',          'Got questions? Ask them here and the community will help.',                     NULL, 0, NULL),
  ('550e8400-e29b-41d4-a716-446655440001', 'Alice''s Group Topic',          'This topic belongs to the Go Devs group.',                                      NULL, 0, '770e8400-e29b-41d4-a716-446655440001');

------------------------------------------------------------
-- COMMENTS (10) — id is INTEGER AUTOINCREMENT
------------------------------------------------------------
INSERT OR IGNORE INTO comments (user_id, topic_id, content, image_path)
VALUES
  ('550e8400-e29b-41d4-a716-446655440002', 1, 'Hey everyone! Excited to be here.',              NULL),
  ('550e8400-e29b-41d4-a716-446655440004', 1, 'Welcome Bob! Glad to have you.',                  NULL),
  ('550e8400-e29b-41d4-a716-446655440001', 2, 'Rust has better memory safety guarantees.',       NULL),
  ('550e8400-e29b-41d4-a716-446655440004', 2, 'Go is simpler to learn though.',                  NULL),
  ('550e8400-e29b-41d4-a716-446655440005', 2, 'I''m just starting to learn both!',               NULL),
  ('550e8400-e29b-41d4-a716-446655440002', 4, 'Great guidelines, very clear.',                   NULL),
  ('550e8400-e29b-41d4-a716-446655440001', 5, 'Count me in for the hackathon!',                  NULL),
  ('550e8400-e29b-41d4-a716-446655440002', 5, 'What are the rules? Any theme?',                 NULL),
  ('550e8400-e29b-41d4-a716-446655440001', 6, 'GitLens is a must-have for any project.',         NULL),
  ('550e8400-e29b-41d4-a716-446655440005', 6, 'Thanks for the recommendations!',                 NULL);

------------------------------------------------------------
-- VOTES (12) — id is INTEGER AUTOINCREMENT, column is reaction_type
------------------------------------------------------------
INSERT OR IGNORE INTO votes (user_id, topic_id, comment_id, reaction_type)
VALUES
  -- Topic votes
  ('550e8400-e29b-41d4-a716-446655440002', 1,  NULL, 1),
  ('550e8400-e29b-41d4-a716-446655440003', 1,  NULL, 1),
  ('550e8400-e29b-41d4-a716-446655440001', 2,  NULL, 1),
  ('550e8400-e29b-41d4-a716-446655440004', 2,  NULL, 1),
  ('550e8400-e29b-41d4-a716-446655440005', 2,  NULL, -1),
  ('550e8400-e29b-41d4-a716-446655440002', 4,  NULL, 1),
  ('550e8400-e29b-41d4-a716-446655440001', 5,  NULL, 1),
  -- Comment votes
  ('550e8400-e29b-41d4-a716-446655440001', NULL, 1, 1),
  ('550e8400-e29b-41d4-a716-446655440004', NULL, 3, 1),
  ('550e8400-e29b-41d4-a716-446655440002', NULL, 4, 1),
  ('550e8400-e29b-41d4-a716-446655440003', NULL, 5, 1),
  ('550e8400-e29b-41d4-a716-446655440001', NULL, 9, 1);

------------------------------------------------------------
-- FOLLOWS (7)
------------------------------------------------------------
INSERT OR IGNORE INTO follows (follower_id, followee_id, created_at)
VALUES
  ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '2025-01-10 09:00:00'),
  ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', '2025-01-11 10:00:00'),
  ('550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', '2025-01-12 11:00:00'),
  ('550e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', '2025-01-13 12:00:00'),
  ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '2025-01-14 13:00:00'),
  ('550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', '2025-01-15 14:00:00'),
  ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440004', '2025-01-16 15:00:00');

------------------------------------------------------------
-- CHATS (2) and MESSAGES (4)
------------------------------------------------------------
INSERT OR IGNORE INTO chats (id, user_one_id, user_two_id, created_at)
VALUES
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '2025-02-01 10:00:00'),
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440005', '2025-02-02 11:00:00');

-- Chat reads (acts as participant tracking)
INSERT OR IGNORE INTO chat_reads (chat_id, user_id, unread_count, updated_at)
VALUES
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 0, '2025-02-01 10:00:00'),
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 0, '2025-02-01 10:00:00'),
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 0, '2025-02-02 11:00:00'),
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440005', 0, '2025-02-02 11:00:00');

-- Messages (id is INTEGER AUTOINCREMENT)
INSERT OR IGNORE INTO messages (chat_id, sender_id, content, created_at, client_message_id)
VALUES
  -- Chat 1: Alice <-> Bob
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Hey Bob, welcome to the forum!',   '2025-02-01 10:05:00', 'seed-msg-001'),
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'Thanks Alice! Happy to be here.', '2025-02-01 10:10:00', 'seed-msg-002'),
  -- Chat 2: Diana <-> Eve
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 'Are you joining the hackathon?',   '2025-02-02 11:05:00', 'seed-msg-003'),
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440005', 'Yes! What''s the theme?',           '2025-02-02 11:10:00', 'seed-msg-004');

------------------------------------------------------------
-- GROUPS (2)
------------------------------------------------------------
INSERT OR IGNORE INTO groups (id, title, description, creator_id)
VALUES
  ('770e8400-e29b-41d4-a716-446655440001', 'Go Devs',          'A group for Go developers to share knowledge and projects.', '550e8400-e29b-41d4-a716-446655440001'),
  ('770e8400-e29b-41d4-a716-446655440002', 'Event Planners',   'Coordinate and plan community events.',                      '550e8400-e29b-41d4-a716-446655440004');

-- Group members
INSERT OR IGNORE INTO group_members (group_id, user_id, role)
VALUES
  ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'creator'),
  ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'member'),
  ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440005', 'member'),
  ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 'creator'),
  ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'member');

-- Group posts (id is TEXT PRIMARY KEY)
INSERT OR IGNORE INTO group_posts (id, group_id, author_id, title, content, image_path)
VALUES
  ('bb0e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Go 1.22 Released!',      'Go 1.22 brings some exciting new features including range-over-func.', NULL),
  ('bb0e8400-e29b-41d4-a716-446655440002', '770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'Need Help with Goroutines', 'How do you handle goroutine lifecycle management in large apps?', NULL),
  ('bb0e8400-e29b-41d4-a716-446655440003', '770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 'Hackathon Planning',       'Let''s finalize the schedule for next week''s hackathon.',          NULL);

-- Group post comments (id is TEXT PRIMARY KEY)
INSERT OR IGNORE INTO group_post_comments (id, post_id, author_id, content, image_path)
VALUES
  ('cc0e8400-e29b-41d4-a716-446655440001', 'bb0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'Great news! Can''t wait to try it.', NULL),
  ('cc0e8400-e29b-41d4-a716-446655440002', 'bb0e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Use context.WithCancel and errgroups.', NULL),
  ('cc0e8400-e29b-41d4-a716-446655440003', 'bb0e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'I''ll handle the opening ceremony.', NULL);

------------------------------------------------------------
-- EVENTS (1) with OPTIONS and RSVPs
------------------------------------------------------------
INSERT OR IGNORE INTO events (id, group_id, creator_id, title, description, event_time)
VALUES
  ('dd0e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 'Community Hackathon 2025', 'A weekend hackathon for building cool projects together!', '2025-03-15 09:00:00');

-- Event options (id is TEXT PRIMARY KEY)
INSERT OR IGNORE INTO event_options (id, event_id, name)
VALUES
  ('ee0e8400-e29b-41d4-a716-446655440001', 'dd0e8400-e29b-41d4-a716-446655440001', 'Saturday Morning'),
  ('ee0e8400-e29b-41d4-a716-446655440002', 'dd0e8400-e29b-41d4-a716-446655440001', 'Saturday Afternoon'),
  ('ee0e8400-e29b-41d4-a716-446655440003', 'dd0e8400-e29b-41d4-a716-446655440001', 'Sunday Morning');

-- Event RSVPs (no status column in schema)
INSERT OR IGNORE INTO event_rsvps (event_id, user_id, option_id)
VALUES
  ('dd0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440004', 'ee0e8400-e29b-41d4-a716-446655440001'),
  ('dd0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'ee0e8400-e29b-41d4-a716-446655440002'),
  ('dd0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'ee0e8400-e29b-41d4-a716-446655440001'),
  ('dd0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440005', 'ee0e8400-e29b-41d4-a716-446655440003');

------------------------------------------------------------
-- NOTIFICATIONS (6) — id is TEXT PRIMARY KEY
------------------------------------------------------------
INSERT OR IGNORE INTO notifications (id, user_id, type, source_id, content, is_read, created_at)
VALUES
  ('ff0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'follow',       '550e8400-e29b-41d4-a716-446655440002', 'Bob started following you.',                          0, '2025-01-10 09:00:00'),
  ('ff0e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'follow',       '550e8400-e29b-41d4-a716-446655440003', 'Charlie started following you.',                      0, '2025-01-11 10:00:00'),
  ('ff0e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'comment',      '880e8400-e29b-41d4-a716-446655440002', 'Diana commented on your topic "Welcome to the Forum!"', 0, '2025-01-20 14:00:00'),
  ('ff0e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', 'comment',      '880e8400-e29b-41d4-a716-446655440003', 'Alice commented on your topic "Go vs Rust: Performance"', 0, '2025-01-21 15:00:00'),
  ('ff0e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440004', 'group_invite', '770e8400-e29b-41d4-a716-446655440001', 'Alice invited you to group "Go Devs".',              0, '2025-01-22 16:00:00'),
  ('ff0e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', 'event',        'dd0e8400-e29b-41d4-a716-446655440001', 'Diana created event "Community Hackathon 2025".',     0, '2025-02-01 12:00:00');

COMMIT;
