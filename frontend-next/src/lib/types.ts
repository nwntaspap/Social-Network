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
  // Support both camelCase (API) and PascalCase (Go backend) field names
  id?: number;
  ID?: number;
  name?: string;
  Name?: string;
  description?: string;
  Description?: string;
  color?: string;
  Color?: string;
  image_path?: string;
  ImagePath?: string;
  topic_count?: number;
  TopicCount?: number;
  topics?: Topic[];
  Topics?: Topic[];
  created_at?: string;
  CreatedAt?: string;
}

export interface Topic {
  // Support both camelCase and PascalCase
  id?: number;
  ID?: number;
  title?: string;
  Title?: string;
  content?: string;
  Content?: string;
  user_id?: number;
  UserID?: number;
  username?: string;
  Username?: string;
  category_id?: number;
  CategoryID?: number;
  created_at?: string;
  CreatedAt?: string;
  updated_at?: string;
  UpdatedAt?: string;
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
