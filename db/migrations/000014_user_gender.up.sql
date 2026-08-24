-- 000014_user_gender.up.sql

ALTER TABLE users ADD COLUMN gender TEXT NOT NULL DEFAULT '';
