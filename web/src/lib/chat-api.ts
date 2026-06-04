import { request } from './client'

// ===== Types =====
export interface Message {
  id: string
  conversation_type: string
  conversation_id: string
  seq_id: number
  sender_id: string
  content_type: string
  content: string
  extra: string | null
  reply_to_id: string | null
  status: string
  created_at: number
}

export interface FriendRequest {
  id: string
  sender_id: string
  receiver_id: string
  status: string
  remark: string
  created_at: number
  updated_at: number
}

export interface Friend {
  id: string
  user_id: string
  friend_id: string
  remark: string
  last_message_at: number
  created_at: number
}

export interface Group {
  id: string
  name: string
  avatar: string
  owner_id: string
  created_at: number
  updated_at: number
  member_count: number
}

export interface GroupMember {
  id: string
  group_id: string
  user_id: string
  role: string
  joined_at: number
}

export interface ConversationInfo {
  conversation_type: string
  conversation_id: string
  name: string
  avatar: string
  last_message: Message | null
  unread_count: number
  last_message_at: number
  pinned: boolean
  mentioned: boolean
}

export interface SearchUserInfo {
  id: string
  name: string
  avatar: string
}

// ===== Friend API =====
export const friendApi = {
  sendRequest: (receiverId: string, remark: string) =>
    request<{ friend_request: FriendRequest }>('/friends/requests', {
      method: 'POST', body: JSON.stringify({ receiver_id: receiverId, remark }),
    }),
  getReceivedRequests: () =>
    request<{ friend_requests: FriendRequest[] }>('/friends/requests/received'),
  getSentRequests: () =>
    request<{ friend_requests: FriendRequest[] }>('/friends/requests/sent'),
  acceptRequest: (requestId: string) =>
    request(`/friends/requests/${requestId}/accept`, { method: 'POST' }),
  rejectRequest: (requestId: string) =>
    request(`/friends/requests/${requestId}/reject`, { method: 'POST' }),
  getList: (query?: string) =>
    request<{ friends: Friend[] }>(`/friends${query ? '?q=' + encodeURIComponent(query) : ''}`),
  deleteFriend: (friendId: string) =>
    request(`/friends/${friendId}`, { method: 'DELETE' }),
  updateRemark: (friendId: string, remark: string) =>
    request(`/friends/${friendId}/remark`, { method: 'PUT', body: JSON.stringify({ remark }) }),
  searchUsers: (q: string) =>
    request<{ users: SearchUserInfo[] }>('/users/search?q=' + encodeURIComponent(q)),
}

// ===== Message API =====
export const messageApi = {
  sendPrivate: (data: { receiver_id: string; content_type: string; content: string; extra?: string; reply_to_id?: string }) =>
    request<{ message: Message }>('/messages/private', { method: 'POST', body: JSON.stringify(data) }),
  sendGroup: (data: { group_id: string; content_type: string; content: string; extra?: string; mentions?: string[]; reply_to_id?: string }) =>
    request<{ message: Message }>('/messages/group', { method: 'POST', body: JSON.stringify(data) }),
  getMessages: (conversationType: string, conversationId: string, beforeSeqId?: number, limit?: number) => {
    let path = `/messages?conversation_type=${conversationType}&conversation_id=${conversationId}`
    if (beforeSeqId) path += `&before_seq_id=${beforeSeqId}`
    if (limit) path += `&limit=${limit}`
    return request<{ messages: Message[] }>(path)
  },
  recallMessage: (messageId: string) =>
    request(`/messages/${messageId}/recall`, { method: 'POST' }),
  forwardMessage: (data: { target_conversation_type: string; target_conversation_id: string; message_ids: string[] }) =>
    request<{ message: Message }>('/messages/forward', { method: 'POST', body: JSON.stringify(data) }),
  getUnreadCount: () =>
    request<{ counts: Record<string, number> }>('/messages/unread-count'),
}

// ===== Group API =====
export const groupApi = {
  create: (data: { name: string; avatar?: string; member_ids?: string[] }) =>
    request<{ group: Group }>('/groups', { method: 'POST', body: JSON.stringify(data) }),
  getInfo: (groupId: string) =>
    request<{ group: Group; is_member: boolean }>(`/groups/${groupId}`),
  update: (groupId: string, data: { name?: string; avatar?: string }) =>
    request(`/groups/${groupId}`, { method: 'PUT', body: JSON.stringify(data) }),
  leave: (groupId: string) =>
    request(`/groups/${groupId}`, { method: 'DELETE' }),
  addMembers: (groupId: string, memberIds: string[]) =>
    request(`/groups/${groupId}/members`, { method: 'POST', body: JSON.stringify({ member_ids: memberIds }) }),
  removeMember: (groupId: string, userId: string) =>
    request(`/groups/${groupId}/members/${userId}`, { method: 'DELETE' }),
  getMembers: (groupId: string) =>
    request<{ members: GroupMember[] }>(`/groups/${groupId}/members`),
  getMyGroups: () =>
    request<{ groups: Group[] }>('/groups'),
  transferOwnership: (groupId: string, newOwnerId: string) =>
    request(`/groups/${groupId}/transfer`, { method: 'POST', body: JSON.stringify({ new_owner_id: newOwnerId }) }),
  setAdmin: (groupId: string, userId: string) =>
    request(`/groups/${groupId}/admins`, { method: 'POST', body: JSON.stringify({ user_id: userId }) }),
  removeAdmin: (groupId: string, userId: string) =>
    request(`/groups/${groupId}/admins/${userId}`, { method: 'DELETE' }),
}

// ===== Conversation API =====
export const conversationApi = {
  getList: () =>
    request<{ conversations: ConversationInfo[] }>('/conversations'),
  markRead: (type: string, id: string) =>
    request(`/conversations/${type}/${id}/read`, { method: 'POST' }),
  togglePin: (type: string, id: string) =>
    request<{ pinned: boolean }>(`/conversations/${type}/${id}/pin`, { method: 'POST' }),
  deleteConversation: (type: string, id: string) =>
    request(`/conversations/${type}/${id}`, { method: 'DELETE' }),
}

