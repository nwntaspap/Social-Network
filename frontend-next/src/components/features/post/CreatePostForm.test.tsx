import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import CreatePostForm from './CreatePostForm';

const mockCreatePost = vi.fn().mockResolvedValue({ id: 't1' });
const mockCreateGroupPost = vi.fn().mockResolvedValue({ id: 'gp1' });

const { useSearchParams, setSearchParams } = vi.hoisted(() => {
  let current = new URLSearchParams();
  return {
    useSearchParams: () => current,
    setSearchParams: (query: string) => {
      current = new URLSearchParams(query);
    },
  };
});

vi.mock('@/lib/api', () => ({
  createPost: (formData: FormData) => mockCreatePost(formData),
  createGroupPost: (formData: FormData, groupId: string) => mockCreateGroupPost(formData, groupId),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  }),
  useSearchParams,
}));

vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    user: {
      id: '1',
      email: 'test@test.com',
      username: 'testuser',
      firstName: 'Test',
      lastName: 'User',
    },
  }),
}));

function typeContent() {
  fireEvent.change(screen.getByPlaceholderText("What's on your mind?"), {
    target: { value: 'A brand new group post' },
  });
}

describe('CreatePostForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setSearchParams('');
  });

  it('calls createGroupPost when a groupId is present', async () => {
    setSearchParams('groupId=g1');

    render(<CreatePostForm />);
    typeContent();
    fireEvent.click(screen.getByRole('button', { name: 'Post' }));

    await waitFor(() => {
      expect(mockCreateGroupPost).toHaveBeenCalledTimes(1);
    });
    const [formData, groupId] = mockCreateGroupPost.mock.calls[0];
    expect(groupId).toBe('g1');
    expect(formData.get('title')).toBe('A brand new group post');
    expect(formData.get('content')).toBe('A brand new group post');
    expect(mockCreatePost).not.toHaveBeenCalled();
  });

  it('calls createPost when no groupId is present', async () => {
    render(<CreatePostForm />);
    typeContent();
    fireEvent.click(screen.getByRole('button', { name: 'Post' }));

    await waitFor(() => {
      expect(mockCreatePost).toHaveBeenCalledTimes(1);
    });
    expect(mockCreateGroupPost).not.toHaveBeenCalled();
  });
});
