'use client';

/**
 * components/features/chat/EmojiPicker.tsx
 *
 * Popover with a small grid of emojis. Clicking one emits it via onPick.
 */

import { useEffect, useRef, useState } from 'react';
import { EMOJI_CATEGORIES } from '@/lib/emoji';

interface EmojiPickerProps {
  onPick: (emoji: string) => void;
  onClose?: () => void;
}

export default function EmojiPicker({ onPick, onClose }: EmojiPickerProps) {
  const [activeCategory, setActiveCategory] = useState(EMOJI_CATEGORIES[0]?.name ?? '');
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose?.();
      }
    }
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [onClose]);

  const category = EMOJI_CATEGORIES.find((c) => c.name === activeCategory) ?? EMOJI_CATEGORIES[0];

  return (
    <div className="emoji-picker" ref={ref}>
      <div className="emoji-picker-tabs">
        {EMOJI_CATEGORIES.map((c) => (
          <button
            key={c.name}
            type="button"
            className={`emoji-picker-tab${c.name === category?.name ? ' emoji-picker-tab--active' : ''}`}
            onClick={() => setActiveCategory(c.name)}
            aria-label={c.name}
          >
            {c.emojis[0]}
          </button>
        ))}
      </div>
      <div className="emoji-picker-grid">
        {category?.emojis.map((emoji) => (
          <button
            key={emoji}
            type="button"
            className="emoji-picker-emoji"
            onClick={() => onPick(emoji)}
          >
            {emoji}
          </button>
        ))}
      </div>
    </div>
  );
}
