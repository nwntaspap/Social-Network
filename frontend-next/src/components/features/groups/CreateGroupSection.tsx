'use client';

import { useState } from 'react';
import CreateGroupForm from './CreateGroupForm';

interface CreateGroupSectionProps {
  onCreated?: () => void;
}

export default function CreateGroupSection({ onCreated }: CreateGroupSectionProps) {
  const [showGroupForm, setShowGroupForm] = useState(false);

  return (
    <div className="create-group-wrapper">
      <button className="create-group-btn" onClick={() => setShowGroupForm(!showGroupForm)}>
        {showGroupForm ? 'Cancel' : 'Create Group'}
      </button>
      {showGroupForm && (
        <CreateGroupForm onClose={() => setShowGroupForm(false)} onCreated={onCreated} />
      )}
    </div>
  );
}
