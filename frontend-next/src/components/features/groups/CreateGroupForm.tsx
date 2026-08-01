'use client';

import { useState } from 'react';
import { createGroup } from '@/lib/api';

interface CreateGroupFormProps {
  onClose: () => void;
  onCreated?: () => void;
}

export default function CreateGroupForm({ onClose, onCreated }: CreateGroupFormProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const [titleError, setTitleError] = useState('');
  const [descriptionError, setDescriptionError] = useState('');
  const [submitError, setSubmitError] = useState('');

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

    setSubmitting(true);
    setSubmitError('');

    try {
      await createGroup(title.trim(), description.trim());
      setTitle('');
      setDescription('');
      onCreated?.();
      onClose();
    } catch {
      setSubmitError('Failed to create group. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="create-group-form">
      <form onSubmit={handleSubmit}>
        <div className="create-group-field">
          <label htmlFor="group-title">Group Title</label>
          <input
            id="group-title"
            type="text"
            className="form-input"
            placeholder="Enter group title"
            value={title}
            onChange={(e) => {
              setTitle(e.target.value);
              setTitleError('');
            }}
          />
          {titleError && <p>{titleError}</p>}
        </div>

        <div className="create-group-field">
          <label htmlFor="group-desc">Description</label>
          <textarea
            id="group-desc"
            className="form-input form-textarea"
            placeholder="What is this group about?"
            rows={3}
            value={description}
            onChange={(e) => {
              setDescription(e.target.value);
              setDescriptionError('');
            }}
          />
          {descriptionError && <p>{descriptionError}</p>}
        </div>

        {submitError && <p className="create-group-error">{submitError}</p>}

        <div className="create-group-actions">
          <button type="button" className="create-group-cancel" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="create-group-submit" disabled={submitting}>
            {submitting ? 'Creating...' : 'Submit Group'}
          </button>
        </div>
      </form>
    </div>
  );
}
