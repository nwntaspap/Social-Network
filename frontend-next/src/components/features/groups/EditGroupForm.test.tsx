import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('@/lib/api', () => ({
  updateGroup: vi.fn().mockResolvedValue({
    id: 'g1',
    title: 'Updated Title',
    description: 'Updated desc',
    creatorId: 'u1',
    creator: {} as never,
    membersCount: 1,
  }),
}));

import { updateGroup } from '@/lib/api';
import EditGroupForm from './EditGroupForm';
import type { Group } from '@/lib/types';

const group: Group = {
  id: 'g1',
  title: 'Go Meetup',
  description: 'Gophers hang out',
  creatorId: 'u1',
  creator: {} as never,
  membersCount: 2,
  createdAt: '2026-01-01T00:00:00Z',
};

describe('EditGroupForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('pre-fills the form with the current group values', () => {
    render(<EditGroupForm group={group} onClose={vi.fn()} />);

    expect(screen.getByLabelText(/Group Title/i)).toHaveValue('Go Meetup');
    expect(screen.getByLabelText(/Description/i)).toHaveValue('Gophers hang out');
  });

  it('calls updateGroup with trimmed values and reports success', async () => {
    const onClose = vi.fn();
    const onUpdated = vi.fn();

    render(<EditGroupForm group={group} onClose={onClose} onUpdated={onUpdated} />);

    fireEvent.change(screen.getByLabelText(/Group Title/i), {
      target: { value: '  Rust Users  ' },
    });
    fireEvent.change(screen.getByLabelText(/Description/i), {
      target: { value: '  Ferris fans  ' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));

    await waitFor(() => {
      expect(updateGroup).toHaveBeenCalledWith('g1', {
        title: 'Rust Users',
        description: 'Ferris fans',
      });
    });
    expect(onUpdated).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('does not call updateGroup when title is empty', () => {
    render(<EditGroupForm group={group} onClose={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/Group Title/i), { target: { value: '  ' } });
    fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));

    expect(updateGroup).not.toHaveBeenCalled();
  });
});
