import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('@/lib/api', () => ({
  createGroup: vi.fn().mockResolvedValue({
    id: 'g1',
    title: 'My Group',
    description: 'About my group',
    creatorId: '1',
    creator: {} as never,
    membersCount: 1,
  }),
}));

import { createGroup } from '@/lib/api';
import CreateGroupForm from './CreateGroupForm';

describe('CreateGroupForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createGroup with the title and description on submit', async () => {
    const onClose = vi.fn();
    const onCreated = vi.fn();

    render(<CreateGroupForm onClose={onClose} onCreated={onCreated} />);

    fireEvent.change(screen.getByLabelText(/Group Title/i), {
      target: { value: 'My Group' },
    });
    fireEvent.change(screen.getByLabelText(/Description/i), {
      target: { value: 'About my group' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Submit Group/i }));

    await waitFor(() => {
      expect(createGroup).toHaveBeenCalledWith('My Group', 'About my group');
    });
    expect(onCreated).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('does not call createGroup when fields are empty', () => {
    render(<CreateGroupForm onClose={vi.fn()} onCreated={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: /Submit Group/i }));

    expect(createGroup).not.toHaveBeenCalled();
  });
});
