-- 000017: Create groups, group_members, group_invitations, group_join_requests,
--         group_posts, group_post_comments tables.

CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    creator_id TEXT NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_groups_creator ON groups(creator_id);

-- group_members: row presence = membership, role column for creator/admin/member
CREATE TABLE group_members (
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('creator', 'admin', 'member')),
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_members_user ON group_members(user_id);

-- group_invitations: row presence = pending invitation
-- accept/decline = delete the row
CREATE TABLE group_invitations (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    inviter_id TEXT NOT NULL REFERENCES users(id),
    invitee_id TEXT NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, invitee_id)
);

CREATE INDEX idx_group_invitations_invitee ON group_invitations(invitee_id);

-- group_join_requests: row presence = pending join request
-- approve/decline = delete the row
CREATE TABLE group_join_requests (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    requester_id TEXT NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, requester_id)
);

CREATE INDEX idx_group_join_requests_group ON group_join_requests(group_id);

-- group_posts: posts inside a group
CREATE TABLE group_posts (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_posts_group ON group_posts(group_id, created_at DESC);
CREATE INDEX idx_group_posts_author ON group_posts(author_id);

-- group_post_comments: comments on group posts
CREATE TABLE group_post_comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL REFERENCES group_posts(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_post_comments_post ON group_post_comments(post_id, created_at ASC);
