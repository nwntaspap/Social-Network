'use client';

/**
 * components/features/groups/CreateEventForm.tsx
 *
 * Simple form to create an event with title and description.
 *
 * TODO: Wire to createEvent API when backend is ready.
 */

import { useState } from 'react';

interface CreateEventFormProps {
  groupId: string;
  onClose: () => void;
}

export default function CreateEventForm({ groupId, onClose }: CreateEventFormProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [eventDate, setEventDate] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const [titleError, setTitleError] = useState('');
  const [descriptionError, setDescriptionError] = useState('');
  const [eventDateError, setEventDateError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!title) {
      setTitleError('Title cannot be empty');
      return;
    }
    if (!description) {
      setDescriptionError('Description cannot be empty');
      return;
    }
    if (!eventDate) {
      setEventDateError('Event Date cannot be empty');
      return;
    }

    setSubmitting(true);

    // TODO: await createEvent(groupId, { title, description, eventDate });
    console.log('create event', { groupId, title, description, eventDate });

    setSubmitting(false);
    setTitle('');
    setDescription('');
    setEventDate('');
    onClose();
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

        <button type="submit" className="group-action-btn" disabled={submitting}>
          {submitting ? 'Creating...' : 'Create Event'}
        </button>
      </form>
    </div>
  );
}
