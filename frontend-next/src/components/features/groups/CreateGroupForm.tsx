'use client';

/**
 * components/features/groups/CreateGroupForm.tsx
 *
 * Simple form to create a group with title and description.
 *
 * TODO: Wire to createGroup API when backend is ready.
 */

import { useState } from 'react';

interface CreateGroupFormProps {
  onClose: () => void;
}

export default function CreateGroupForm({ onClose }: CreateGroupFormProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const [titleError, setTitleError] = useState('');
  const [descriptionError, setDescriptionError] = useState('');

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

    // TODO: await createGroup(title, description);
    console.log('create group', { title, description });

    setSubmitting(false);
    setTitle('');
    setDescription('');
    onClose();
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
