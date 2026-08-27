'use client';

/**
 * components/features/chat/MessageInput.tsx
 *
 * Textarea + send button with an emoji picker toggle.
 * Submits on Enter (Shift+Enter inserts a newline).
 */

import { FormEvent, useRef, useState } from 'react';
import EmojiPicker from './EmojiPicker';

interface MessageInputProps {
  onSend: (content: string) => void;
  onTyping?: () => void;
  disabled?: boolean;
}

export default function MessageInput({ onSend, onTyping, disabled }: MessageInputProps) {
  const [content, setContent] = useState('');
  const [pickerOpen, setPickerOpen] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setContent('');
    if (textareaRef.current) textareaRef.current.style.height = 'auto';
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const form = (e.target as HTMLTextAreaElement).closest('form');
      form?.requestSubmit();
    }
  }

  function handleChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    setContent(e.target.value);
    e.target.style.height = 'auto';
    e.target.style.height = `${Math.min(e.target.scrollHeight, 100)}px`;
    onTyping?.();
  }

  return (
    <form className="chat-input-area" onSubmit={handleSubmit}>
      <div className="chat-input-emoji-wrap">
        <button
          type="button"
          className="chat-emoji-button"
          aria-label="Pick an emoji"
          onClick={() => setPickerOpen((prev) => !prev)}
        >
          😊
        </button>
        {pickerOpen && (
          <EmojiPicker
            onPick={(emoji) => {
              setContent((prev) => prev + emoji);
              setPickerOpen(false);
              textareaRef.current?.focus();
            }}
            onClose={() => setPickerOpen(false)}
          />
        )}
      </div>
      <textarea
        ref={textareaRef}
        className="chat-input"
        placeholder="Type a message..."
        value={content}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        rows={1}
        disabled={disabled}
      />
      <button type="submit" className="chat-send-button" disabled={disabled || !content.trim()}>
        ➤
      </button>
    </form>
  );
}
