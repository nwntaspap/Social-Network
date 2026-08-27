INSERT INTO group_posts (id, group_id, author_id, title, content, image_path, created_at, updated_at)
SELECT lower(hex(randomblob(16))), t.group_id, t.user_id, t.title, t.content, t.image_path, t.created_at, t.updated_at
FROM topics t
WHERE t.group_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM groups g WHERE g.id = t.group_id)
  AND NOT EXISTS (SELECT 1 FROM group_posts gp WHERE gp.group_id = t.group_id AND gp.title = t.title AND gp.content = t.content);
