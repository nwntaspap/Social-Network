import { describe, it, expect } from 'vitest';
import {
  getNotificationMessage,
  getNotificationHref,
  hasNotificationActions,
  getNotificationIcon,
} from './notificationFormat';
import type { Notification } from './types';

function makeNotification(overrides: Partial<Notification> = {}): Notification {
  return {
    id: 1,
    recipient_id: 'u2',
    type: 'like',
    resource_type: 'post',
    resource_id: '42',
    actor_id: 'u1',
    actor_name: 'Alice',
    actor_avatar: '',
    content_text: '',
    image_url: '',
    join_request_id: '',
    event_id: '',
    is_read: false,
    created_at: '2026-08-14T10:00:00Z',
    deleted: false,
    ...overrides,
  };
}

describe('getNotificationMessage', () => {
  it('formats like on a post', () => {
    expect(getNotificationMessage(makeNotification({ type: 'like' }))).toBe(
      'Alice liked your post'
    );
  });

  it('formats like on a comment', () => {
    expect(
      getNotificationMessage(makeNotification({ type: 'like', resource_type: 'comment' }))
    ).toBe('Alice liked your comment');
  });

  it('formats dislike', () => {
    expect(getNotificationMessage(makeNotification({ type: 'dislike' }))).toBe(
      'Alice disliked your post'
    );
  });

  it('formats follow', () => {
    expect(getNotificationMessage(makeNotification({ type: 'follow' }))).toBe('Alice followed you');
  });

  it('formats follow_request', () => {
    expect(getNotificationMessage(makeNotification({ type: 'follow_request' }))).toBe(
      'Alice wants to follow you'
    );
  });

  it('formats group invite', () => {
    expect(getNotificationMessage(makeNotification({ type: 'group_invite' }))).toBe(
      'Alice invited you to join a group'
    );
  });

  it('formats group join request', () => {
    expect(getNotificationMessage(makeNotification({ type: 'group_join_request' }))).toBe(
      'Alice wants to join your group'
    );
  });

  it('formats group join accept', () => {
    expect(getNotificationMessage(makeNotification({ type: 'group_join_accept' }))).toBe(
      'You joined the group'
    );
  });

  it('formats group join accept with the group title', () => {
    expect(
      getNotificationMessage(makeNotification({ type: 'group_join_accept', content_text: 'Devs' }))
    ).toBe('You joined group "Devs"');
  });

  it('formats pending group invite acceptance', () => {
    expect(getNotificationMessage(makeNotification({ type: 'group_invite_pending' }))).toBe(
      'You accepted to join the group'
    );
  });

  it('formats pending group invite acceptance with the group title', () => {
    expect(
      getNotificationMessage(
        makeNotification({ type: 'group_invite_pending', content_text: 'Devs' })
      )
    ).toBe('You accepted to join group "Devs"');
  });

  it('falls back to actor name when actor_name is missing', () => {
    expect(getNotificationMessage(makeNotification({ type: 'like', actor_name: '' }))).toBe(
      'Someone liked your post'
    );
  });

  it('falls back to content_text for unknown types', () => {
    expect(
      getNotificationMessage(
        makeNotification({ type: 'custom_event', content_text: 'An event happened' })
      )
    ).toBe('An event happened');
  });
});

describe('getNotificationHref', () => {
  it('points likes on posts at the post', () => {
    expect(getNotificationHref(makeNotification({ type: 'like' }))).toBe('/post/42');
  });

  it('returns null for likes on comments (resolved via getComment)', () => {
    expect(
      getNotificationHref(makeNotification({ type: 'like', resource_type: 'comment' }))
    ).toBeNull();
  });

  it('points comments at the post', () => {
    expect(getNotificationHref(makeNotification({ type: 'comment' }))).toBe('/post/42');
  });

  it('points follow at the actor profile', () => {
    expect(getNotificationHref(makeNotification({ type: 'follow' }))).toBe('/profile/u1');
  });

  it('points group notifications at the group', () => {
    expect(getNotificationHref(makeNotification({ type: 'group_invite', resource_id: 'g1' }))).toBe(
      '/groups/g1'
    );
    expect(
      getNotificationHref(makeNotification({ type: 'group_invite_pending', resource_id: 'g1' }))
    ).toBe('/groups/g1');
  });

  it('returns null when a group notification has no resource_id', () => {
    expect(
      getNotificationHref(makeNotification({ type: 'group_invite', resource_id: '' }))
    ).toBeNull();
  });
});

describe('hasNotificationActions', () => {
  it.each(['follow_request', 'group_invite', 'group_join_request'])(
    'returns true for %s',
    (type) => {
      expect(hasNotificationActions(makeNotification({ type }))).toBe(true);
    }
  );

  it.each(['like', 'dislike', 'follow', 'comment', 'group_invite_pending'])(
    'returns false for %s',
    (type) => {
      expect(hasNotificationActions(makeNotification({ type }))).toBe(false);
    }
  );
});

describe('getNotificationIcon', () => {
  it('maps like to like', () => {
    expect(getNotificationIcon(makeNotification({ type: 'like' }))).toBe('like');
  });

  it('maps dislike to dislike', () => {
    expect(getNotificationIcon(makeNotification({ type: 'dislike' }))).toBe('dislike');
  });

  it('maps follow types to follow', () => {
    expect(getNotificationIcon(makeNotification({ type: 'follow' }))).toBe('follow');
    expect(getNotificationIcon(makeNotification({ type: 'follow_request' }))).toBe('follow');
  });

  it('maps group types to group', () => {
    expect(getNotificationIcon(makeNotification({ type: 'group_invite' }))).toBe('group');
    expect(getNotificationIcon(makeNotification({ type: 'group_join_request' }))).toBe('group');
    expect(getNotificationIcon(makeNotification({ type: 'group_invite_pending' }))).toBe('group');
  });

  it('defaults to mention', () => {
    expect(getNotificationIcon(makeNotification({ type: 'custom' }))).toBe('mention');
  });
});
