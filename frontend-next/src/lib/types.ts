// ─── Types ────────────────────────────────────────────────────────────────────

export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  created_at: string;
  // Extend with your actual user fields
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  topics?: Topic[];
  created_at: string;
}

export interface Topic {
  id: number;
  title: string;
  content: string;
  user_id: number;
  username?: string;
  category_id?: number;
  created_at: string;
  updated_at?: string;
}

export interface Comment {
  id: number;
  content: string;
  user_id: number;
  username?: string;
  topic_id: number;
  created_at: string;
}

export interface ChatUser {
  id: number;
  username: string;
  avatar_url?: string;
  is_online: boolean;
  last_message_at?: string;
}

export interface Chat {
  chat_id: number;
  user_id: number;
  other_user: ChatUser;
  created_at: string;
}
