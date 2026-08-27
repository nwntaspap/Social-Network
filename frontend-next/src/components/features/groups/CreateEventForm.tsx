'use client';

import { useState } from 'react';
import { createEvent } from '@/lib/api';
import { ApiError } from '@/lib/api';

interface CreateEventFormProps {
  groupId: string;
  onClose: () => void;
  onCreated?: () => void;
}

export default function CreateEventForm({ groupId, onClose, onCreated }: CreateEventFormProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [eventDate, setEventDate] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const [titleError, setTitleError] = useState('');
  const [descriptionError, setDescriptionError] = useState('');
  const [eventDateError, setEventDateError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!title.trim()) {
      setTitleError('Title cannot be empty');
      return;
    }
    if (!description.trim()) {
      setDescriptionError('Description cannot be empty');
      return;
    }
    if (!eventDate) {
      setEventDateError('Event Date cannot be empty');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      await createEvent(groupId, {
        title,
        description,
        eventDate: new Date(eventDate).toISOString(),
        options: ['Going', 'Not going'],
      });
      setTitle('');
      setDescription('');
      setEventDate('');
      onClose();
      onCreated?.();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to create event. Please try again.');
      setSubmitting(false);
    }
  }

  return (
    <div className="group-dropdown details-form">
      <div className="group-dropdown-header">
        <span>Create Event</span>
        <button type="button" className="group-dropdown-close" onClick={onClose}>
          ✕
        </button>
      </div>

      <form className="event-form" onSubmit={handleSubmit}>
        <div className="event-form-field">
          <label htmlFor="event-title">Title</label>
          <input
            id="event-title"
            type="text"
            className="form-input"
            placeholder="Event title"
            value={title}
            onChange={(e) => {
              setTitle(e.target.value);
              setTitleError('');
            }}
          />
          {titleError && <p>{titleError}</p>}
        </div>

        <div className="event-form-field">
          <label htmlFor="event-desc">Description</label>
          <textarea
            id="event-desc"
            className="form-input form-textarea"
            placeholder="Event description"
            rows={3}
            value={description}
            onChange={(e) => {
              setDescription(e.target.value);
              setDescriptionError('');
            }}
          />
          {descriptionError && <p>{descriptionError}</p>}
        </div>

        <div className="event-form-field">
          <label htmlFor="event-date">Date & Time</label>
          <input
            id="event-date"
            type="datetime-local"
            className="form-input"
            value={eventDate}
            onChange={(e) => {
              setEventDate(e.target.value);
              setEventDateError('');
            }}
          />
          {eventDateError && <p>{eventDateError}</p>}
        </div>

        {error && <p className="create-post-error">{error}</p>}

        <button type="submit" className="group-action-btn" disabled={submitting}>
          {submitting ? 'Creating...' : 'Create Event'}
        </button>
      </form>
    </div>
  );
}
