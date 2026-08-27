// src/lib/types.ts

export interface User {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  dateOfBirth: string;
  avatarUrl?: string;
  nickname?: string;
  aboutMe?: string;
  gender?: string;
  isPublic: boolean;
  isOnline?: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface Profile extends User {
  followersCount?: number;
  followingCount?: number;
  postsCount?: number;
  isFollowing?: boolean;
  isPending?: boolean;
}

export type PostPrivacy = 'public' | 'followers' | 'private';

export interface Post {
  id: string;
  userId: string;
  user: User;
  title?: string;
  content: string;
  imageUrl?: string;
  gifUrl?: string;
  privacy: PostPrivacy;
  allowedUsers?: string[];
  groupId?: string; // if post belongs to a group
  group?: Pick<Group, 'id' | 'title'>; // group info for display
  commentsCount: number;
  likesCount: number;
  dislikesCount?: number;
  isLiked?: boolean;
  userVote?: number | null;
  createdAt: string;
  updatedAt?: string;
}

export interface Comment {
  id: string;
  postId: string;
  userId: string;
  user: User;
  content: string;
  imageUrl?: string;
  gifUrl?: string;
  upvoteCount?: number;
  downvoteCount?: number;
  voteScore?: number;
  userVote?: number | null;
  createdAt: string;
  updatedAt?: string;
}

export interface FollowRequest {
  id: string;
  requesterId: string;
  requester: User;
  targetId: string;
  target: User;
  status: 'pending' | 'accepted' | 'declined';
  createdAt: string;
}
export type MembershipStatus = 'none' | 'pending' | 'member';
export interface Group {
  id: string;
  title: string;
  description: string;
  creatorId: string;
  creator: User;
  membersCount: number;
  membershipStatus?: MembershipStatus;
  unreadCount?: number;
  createdAt: string;
  updatedAt?: string;
}

export interface GroupMember {
  groupId: string;
  userId: string;
  user: User;
  role: 'creator' | 'admin' | 'member';
  joinedAt: string;
}

export interface GroupInvitation {
  id: string;
  groupId: string;
  group: Pick<Group, 'id' | 'title'>;
  inviterId: string;
  inviter: User;
  inviteeId: string;
  invitee: User;
  status?: 'pending' | 'accepted' | 'declined';
  createdAt: string;
}

export interface GroupJoinRequest {
  id: string;
  groupId: string;
  group: Group;
  requesterId: string;
  requester: User;
  createdAt: string;
}

export interface EventOption {
  id: string;
  label: string;
  tally: number;
}

export interface Event {
  id: string;
  groupId: string;
  creatorId: string;
  creator: User;
  title: string;
  description: string;
  eventDate: string;
  createdAt: string;
  options: EventOption[];
}

export interface GroupEventsResponse {
  events: Event[];
  nextCursor?: string;
}

export interface EventRSVPOption {
  optionId: string;
  optionLabel: string;
  users: User[];
}

export interface EventRSVPsResponse {
  options: EventRSVPOption[];
}

export interface EventResponse {
  id: string;
  eventId: string;
  userId: string;
  user: User;
  response: 'going' | 'notGoing';
  createdAt: string;
}

export interface ChatMessage {
  id: string;
  senderId: string;
  sender: User;
  receiverId?: string;
  groupId?: string;
  content: string;
  type: 'private' | 'group';
  createdAt: string;
}

export interface Chat {
  id: string;
  participants: ChatUser[];
  lastMessage?: ChatMessage;
  unreadCount: number;
  createdAt: string;
}

export interface ChatUser {
  id: string;
  username: string;
  avatarUrl?: string;
  isOnline: boolean;
  lastMessageAt?: string;
}

// ─── Realtime (WebSocket) ─────────────────────────────────────────────────────

/** Wire envelope for every WebSocket message (matches backend realtime.Envelope). */
export interface WsEnvelope {
  type: string;
  request_id?: string;
  payload: unknown;
}

/** Private chat message as pushed over the WebSocket (snake_case from backend). */
export interface PrivateWsMessage {
  id: number;
  chat_id: string;
  sender_id: string;
  content: string;
  created_at: string;
  client_message_id?: string;
}

/** Group chat message as pushed over the WebSocket / returned by HTTP history. */
export interface GroupChatMessageWire {
  id: string;
  group_id: string;
  sender_id: string;
  content: string;
  created_at: string;
}

/** Payload of the isOnlineStatus.update broadcast. */
export interface IsOnlineStatusPayload {
  user_id: string;
  isOnline: boolean;
}

/** Per-member online flag returned by the group presence endpoint. */
export interface GroupPresenceMember {
  id: string;
  isOnline: boolean;
}

/** Live "how many members are online" snapshot for a group. */
export interface GroupPresence {
  groupId: string;
  total: number;
  online: number;
  members: GroupPresenceMember[];
}

/** Payload of the chat.is_typing broadcast. */
export interface IsTypingPayload {
  chat_id: string;
  user_id: string;
}

/** Payload of an error envelope. */
export interface WsErrorPayload {
  message: string;
}

export interface WsConversation extends Chat {
  id: string;
  participants: ChatUser[];
  unreadCount: number;
  createdAt: string;
}

export type NotificationType =
  | 'like'
  | 'dislike'
  | 'follow'
  | 'follow_request'
  | 'follow_accept'
  | 'follow_declined'
  | 'group_invite'
  | 'group_invite_removed'
  | 'group_join_request'
  | 'group_join_accept'
  | 'group_join_declined'
  | 'group_invite_accepted'
  | 'group_invite_declined'
  | 'event'
  | 'post'
  | 'comment'
  | string;

export interface Notification {
  id: number;
  recipient_id: string;
  type: NotificationType;
  resource_type: string;
  resource_id: string;
  actor_id: string;
  actor_name: string;
  actor_avatar: string;
  content_text: string;
  image_url: string;
  join_request_id: string;
  event_id: string;
  is_read: boolean;
  created_at: string;
  deleted: boolean;
}

export interface NotificationsResponse {
  notifications: Notification[];
  total: number;
}

// API Response types
export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  error?: string;
}

export interface GroupPost {
  id: string;
  groupId: string;
  authorId: string;
  author: User;
  title: string;
  content: string;
  imagePath?: string;
  commentsCount: number;
  createdAt: string;
  updatedAt?: string;
}

export interface GroupPostComment {
  id: string;
  postId: string;
  authorId: string;
  author: User;
  content: string;
  imagePath?: string;
  createdAt: string;
}
