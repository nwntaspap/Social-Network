-- 000012_chat_pair_unique.up.sql
--
-- chats(user_one_id, user_two_id) is normalized (low id first), so the pair is
-- unique by design. GetOrCreateChat relies on INSERT OR IGNORE to reuse an
-- existing 1:1 chat, which only works if the DB enforces that uniqueness.
-- Deduplicate any chats created before this constraint existed (keeping the
-- chat with the most messages per pair), then add the unique index.
DELETE FROM chats
WHERE rowid NOT IN (
    SELECT keep_rowid FROM (
        SELECT c.rowid AS keep_rowid,
               ROW_NUMBER() OVER (
                   PARTITION BY c.user_one_id, c.user_two_id
                   ORDER BY (SELECT COUNT(*) FROM messages m WHERE m.chat_id = c.id) DESC, c.created_at ASC, c.rowid ASC
               ) AS rn
        FROM chats c
    ) ranked
    WHERE rn = 1
);

CREATE UNIQUE INDEX idx_chats_pair ON chats (user_one_id, user_two_id);
