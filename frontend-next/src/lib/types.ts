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
  isPublic: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface Profile extends User {
  followersCount: number;
  followingCount: number;
  postsCount: number;
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
  isLiked?: boolean;
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
  group: Group;
  inviterId: string;
  inviter: User;
  inviteeId: string;
  invitee: User;
  status: 'pending' | 'accepted' | 'declined';
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

export interface Event {
  id: string;
  groupId: string;
  creatorId: string;
  creator: User;
  title: string;
  description: string;
  eventDate: string;
  createdAt: string;
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

export type NotificationType =
  | 'followRequest'
  | 'followAccepted'
  | 'groupInvitation'
  | 'groupJoinRequest'
  | 'groupJoinAccepted'
  | 'newEvent'
  | 'newFollower';

export interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  message: string;
  referenceId?: string;
  isRead: boolean;
  createdAt: string;
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
