import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

const { mockCreateEvent } = vi.hoisted(() => ({ mockCreateEvent: vi.fn() }));

vi.mock('@/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api')>();
  return {
    ...actual,
    createEvent: mockCreateEvent,
  };
});

import { createEvent, ApiError } from '@/lib/api';
import CreateEventForm from './CreateEventForm';

describe('CreateEventForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function fillForm() {
    fireEvent.change(screen.getByLabelText(/Title/i), { target: { value: 'Go meetup' } });
    fireEvent.change(screen.getByLabelText(/Description/i), {
      target: { value: 'Monthly meetup' },
    });
    fireEvent.change(screen.getByLabelText(/Date & Time/i), {
      target: { value: '2026-09-01T18:00' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Create Event/i }));
  }

  it('creates an event with the default options', async () => {
    mockCreateEvent.mockResolvedValue({ id: 'evt1' });
    const onClose = vi.fn();
    const onCreated = vi.fn();

    render(<CreateEventForm groupId="g1" onClose={onClose} onCreated={onCreated} />);
    fillForm();

    await waitFor(() => {
      expect(createEvent).toHaveBeenCalledWith('g1', {
        title: 'Go meetup',
        description: 'Monthly meetup',
        eventDate: new Date('2026-09-01T18:00').toISOString(),
        options: ['Going', 'Not going'],
      });
    });
    expect(onCreated).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('shows the real backend error message when creation fails', async () => {
    mockCreateEvent.mockRejectedValue(new ApiError(400, 'at least 2 options are required'));
    const onClose = vi.fn();

    render(<CreateEventForm groupId="g1" onClose={onClose} />);
    fillForm();

    expect(await screen.findByText('at least 2 options are required')).toBeTruthy();
    expect(onClose).not.toHaveBeenCalled();
  });

  it('does not submit when fields are empty', () => {
    render(<CreateEventForm groupId="g1" onClose={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: /Create Event/i }));

    expect(createEvent).not.toHaveBeenCalled();
    expect(screen.getByText('Title cannot be empty')).toBeTruthy();
  });
});
