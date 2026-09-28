import { test, expect } from '@playwright/test';
import { getNotificationHref } from '../src/lib/notificationFormat';
import type { Notification } from '../src/lib/types';

/**
 * E2E validation for: "comment notification for group posts should go to
 * the specific group page (/groups/[id]), not /post/[uuid] (404)".
 *
 * Bug: backend published {type:comment, resource_type:post, resource_id:<group-post-UUID>}
 * with no group_id, frontend built `/post/<uuid>` which expects a numeric
 * topic id -> 400/404 ("Post not found").
 *
 * Run (frontend only, no server needed for contract part):
 *   npx playwright test e2e/group-comment-notification.spec.ts
 *
 * Full manual validation (needs backend + frontend running):
 *   1. User A creates group G, creates post P in G.
 *   2. User B (member) comments on P.
 *   3. User A opens notification bell -> sees "X commented on your post".
 *   4. Click -> expect URL `/groups/<G>` with posts tab, NOT `/post/<P-uuid>` 404.
 */

function makeNotification(overrides: Partial<Notification> = {}): Notification {
  return {
    id: 1,
    recipient_id: 'u-author',
    type: 'comment',
    resource_type: 'post',
    resource_id: 'bb0e8400-1111-2222-3333-444455556666',
    actor_id: 'u-commenter',
    actor_name: 'Bob',
    actor_avatar: '',
    content_text: 'hello',
    image_url: '',
    join_request_id: '',
    event_id: '',
    is_read: false,
    created_at: '2026-08-14T10:00:00Z',
    deleted: false,
    ...overrides,
  } as Notification;
}

test.describe('group comment notification routing (contract)', () => {
  test('normal post comment still goes to /post/<id>', () => {
    const href = getNotificationHref(makeNotification({ resource_id: '42' }));
    expect(href).toBe('/post/42');
  });

  test('group post comment goes to /groups/<groupId>, never /post/<uuid>', () => {
    const href = getNotificationHref(
      makeNotification({
        resource_id: 'bb0e8400-1111-2222-3333-444455556666',
        group_id: 'group-uuid-123',
      } as Partial<Notification>)
    );
    expect(href).toBe('/groups/group-uuid-123');
    expect(href).not.toContain('/post/');
  });

  test('group post like also goes to /groups/<groupId>', () => {
    const href = getNotificationHref(
      makeNotification({
        type: 'like',
        resource_id: 'bb0e8400-1111-2222-3333-444455556666',
        group_id: 'group-uuid-123',
      } as Partial<Notification>)
    );
    expect(href).toBe('/groups/group-uuid-123');
  });

  test('legacy group row (UUID, no group_id) returns null instead of 404 /post/<uuid>', () => {
    const href = getNotificationHref(
      makeNotification({
        resource_id: 'bb0e8400-1111-2222-3333-444455556666',
      } as Partial<Notification>)
    );
    expect(href).toBeNull();
  });
});

test.describe('group page smoke (needs dev server)', () => {
  test.skip(({ page }) => page.url() === 'never', 'template for live run');
});
