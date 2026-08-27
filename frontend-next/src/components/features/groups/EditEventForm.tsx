'use client';

import { useState } from 'react';
import { updateEvent, ApiError } from '@/lib/api';
import type { Event } from '@/lib/types';

interface EditEventFormProps {
  groupId: string;
  event: Event;
  onClose: () => void;
  onUpdated: () => void;
}

function toDatetimeLocal(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '';
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}`;
}

export default function EditEventForm({ groupId, event, onClose, onUpdated }: EditEventFormProps) {
  const [title, setTitle] = useState(event.title);
  const [description, setDescription] = useState(event.description);
  const [eventDate, setEventDate] = useState(toDatetimeLocal(event.eventDate));
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
      await updateEvent(groupId, event.id, {
        title,
        description,
        eventDate: new Date(eventDate).toISOString(),
      });
      onClose();
      onUpdated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update event. Please try again.');
      setSubmitting(false);
    }
  }

  return (
    <div className="group-dropdown details-form">
      <div className="group-dropdown-header">
        <span>Edit Event</span>
        <button type="button" className="group-dropdown-close" onClick={onClose}>
          ✕
        </button>
      </div>

      <form className="event-form" onSubmit={handleSubmit}>
        <div className="event-form-field">
          <label htmlFor="edit-event-title">Title</label>
          <input
            id="edit-event-title"
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
          <label htmlFor="edit-event-desc">Description</label>
          <textarea
            id="edit-event-desc"
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
          <label htmlFor="edit-event-date">Date & Time</label>
          <input
            id="edit-event-date"
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
          {submitting ? 'Saving...' : 'Save Changes'}
        </button>
      </form>
    </div>
  );
}
