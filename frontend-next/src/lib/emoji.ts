/**
 * lib/emoji.ts
 *
 * Curated emoji list used by the chat EmojiPicker. Grouped by category.
 */

export interface EmojiCategory {
  name: string;
  emojis: string[];
}

export const EMOJI_CATEGORIES: EmojiCategory[] = [
  {
    name: 'Smileys',
    emojis: [
      '😀',
      '😄',
      '😁',
      '😆',
      '😂',
      '🤣',
      '😊',
      '😇',
      '🙂',
      '😉',
      '😍',
      '🥰',
      '😘',
      '😎',
      '🤩',
      '😋',
    ],
  },
  {
    name: 'Gestures',
    emojis: [
      '👍',
      '👎',
      '👌',
      '✌️',
      '🤞',
      '🤝',
      '👏',
      '🙌',
      '🙏',
      '💪',
      '🤙',
      '👋',
      '🖐️',
      '☝️',
      '✊',
      '👊',
    ],
  },
  {
    name: 'Hearts',
    emojis: [
      '❤️',
      '🧡',
      '💛',
      '💚',
      '💙',
      '💜',
      '🖤',
      '🤍',
      '💔',
      '💕',
      '💞',
      '💓',
      '💗',
      '💖',
      '💘',
      '❤️‍🔥',
    ],
  },
  {
    name: 'Symbols',
    emojis: [
      '⭐',
      '🌟',
      '✨',
      '⚡',
      '🔥',
      '💯',
      '✅',
      '❌',
      '❗',
      '❓',
      '💬',
      '📣',
      '🎉',
      '🎊',
      '🏆',
      '🚀',
    ],
  },
  {
    name: 'Objects',
    emojis: [
      '🍕',
      '🍔',
      '🌮',
      '🍩',
      '☕',
      '🍺',
      '🎂',
      '🍎',
      '⚽',
      '🎮',
      '🎧',
      '📷',
      '💻',
      '📚',
      '🎁',
      '💡',
    ],
  },
];

export const ALL_EMOJIS: string[] = EMOJI_CATEGORIES.flatMap((c) => c.emojis);
