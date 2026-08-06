import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

const { mockUpdateEvent } = vi.hoisted(() => ({ mockUpdateEvent: vi.fn() }));

vi.mock('@/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api')>();
  return {
    ...actual,
    updateEvent: mockUpdateEvent,
  };
});

import { updateEvent, ApiError } from '@/lib/api';
import EditEventForm from './EditEventForm';

const event = {
  id: 'evt1',
  groupId: 'g1',
  creatorId: 'u1',
  creator: {
    id: 'u1',
    email: 'alice@example.com',
    username: 'alice',
    firstName: 'Alice',
    lastName: 'Doe',
    dateOfBirth: '1995-01-01',
    isPublic: true,
    createdAt: '2024-01-01T00:00:00Z',
  },
  title: 'React meetup',
  description: 'A meetup about React',
  eventDate: '2026-09-01T18:00:00.000Z',
  createdAt: '2026-08-01T10:00:00.000Z',
  options: [
    { id: 'opt1', label: 'Going', tally: 2 },
    { id: 'opt2', label: 'Not going', tally: 1 },
  ],
};

describe('EditEventForm', () => {
  function pad(n: number): string {
    return String(n).padStart(2, '0');
  }

  function toLocalInput(iso: string): string {
    const d = new Date(iso);
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(
      d.getHours()
    )}:${pad(d.getMinutes())}`;
  }

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('pre-fills the form with the event data', () => {
    render(<EditEventForm groupId="g1" event={event} onClose={vi.fn()} onUpdated={vi.fn()} />);

    expect((screen.getByLabelText(/Title/i) as HTMLInputElement).value).toBe('React meetup');
    expect((screen.getByLabelText(/Description/i) as HTMLTextAreaElement).value).toBe(
      'A meetup about React'
    );
    expect((screen.getByLabelText(/Date & Time/i) as HTMLInputElement).value).toBe(
      toLocalInput(event.eventDate)
    );
  });

  it('sends the updated event via updateEvent', async () => {
    mockUpdateEvent.mockResolvedValue({ ...event, title: 'React 20 meetup' });
    const onClose = vi.fn();
    const onUpdated = vi.fn();

    render(<EditEventForm groupId="g1" event={event} onClose={onClose} onUpdated={onUpdated} />);

    fireEvent.change(screen.getByLabelText(/Title/i), { target: { value: 'React 20 meetup' } });
    fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));

    await waitFor(() => {
      expect(updateEvent).toHaveBeenCalledWith('g1', 'evt1', {
        title: 'React 20 meetup',
        description: 'A meetup about React',
        eventDate: new Date(event.eventDate).toISOString(),
      });
    });
    expect(onUpdated).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('shows the real backend error message when the update fails', async () => {
    mockUpdateEvent.mockRejectedValue(new ApiError(403, 'only the group creator can edit events'));
    const onClose = vi.fn();

    render(<EditEventForm groupId="g1" event={event} onClose={onClose} onUpdated={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));

    expect(await screen.findByText('only the group creator can edit events')).toBeTruthy();
    expect(onClose).not.toHaveBeenCalled();
  });
});
