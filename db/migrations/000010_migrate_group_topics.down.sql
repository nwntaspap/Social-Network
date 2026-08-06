DELETE FROM group_posts
WHERE id IN (
    SELECT gp.id
    FROM group_posts gp
    JOIN topics t
      ON t.group_id = gp.group_id
     AND t.user_id = gp.author_id
     AND t.title = gp.title
     AND t.content = gp.content
     AND t.created_at = gp.created_at
    WHERE t.group_id IS NOT NULL
);
