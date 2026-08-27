'use client';

import { useState } from 'react';
import { updateGroup } from '@/lib/api';
import type { Group } from '@/lib/types';

interface EditGroupFormProps {
  group: Group;
  onClose: () => void;
  onUpdated?: () => void;
}

export default function EditGroupForm({ group, onClose, onUpdated }: EditGroupFormProps) {
  const [title, setTitle] = useState(group.title);
  const [description, setDescription] = useState(group.description);
  const [submitting, setSubmitting] = useState(false);
  const [titleError, setTitleError] = useState('');
  const [submitError, setSubmitError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!title.trim()) {
      setTitleError('Title cannot be empty');
      return;
    }

    setSubmitting(true);
    setSubmitError('');

    try {
      await updateGroup(group.id, { title: title.trim(), description: description.trim() });
      onUpdated?.();
      onClose();
    } catch {
      setSubmitError('Failed to update group. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="group-dropdown details-form">
      <div className="group-dropdown-header">
        <span>Edit Group</span>
        <button type="button" className="group-dropdown-close" onClick={onClose}>
          ✕
        </button>
      </div>

      <form className="event-form" onSubmit={handleSubmit}>
        <div className="event-form-field">
          <label htmlFor="edit-group-title">Group Title</label>
          <input
            id="edit-group-title"
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

        <div className="event-form-field">
          <label htmlFor="edit-group-desc">Description</label>
          <textarea
            id="edit-group-desc"
            className="form-input form-textarea"
            placeholder="What is this group about?"
            rows={3}
            value={description}
            onChange={(e) => {
              setDescription(e.target.value);
              setTitleError('');
            }}
          />
        </div>

        {submitError && <p className="create-group-error">{submitError}</p>}

        <div className="create-group-actions">
          <button type="button" className="create-group-cancel" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="create-group-submit" disabled={submitting}>
            {submitting ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
}
