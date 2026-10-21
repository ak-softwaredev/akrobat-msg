const API_BASE = '/api/v1'

function getToken() {
  return localStorage.getItem('token')
}

export function setToken(token) {
  if (token) localStorage.setItem('token', token)
  else localStorage.removeItem('token')
}

async function request(path, options = {}) {
  const token = getToken()
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })

  if (res.status === 204) return null

  const data = await res.json().catch(() => null)

  if (!res.ok) {
    const message =
      data?.error?.message || data?.error || `Request failed (${res.status})`
    throw new Error(message)
  }

  return data
}

export const api = {
  // Auth
  register: (username, password) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  login: (username, password) =>
    request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  logout: () => request('/auth/logout', { method: 'POST' }),

  // Users
  getMe: () => request('/users/me'),
  getUser: (id) => request(`/users/${id}`),
  updateMe: (field, value) =>
    request('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({ field, value }),
    }),
  deleteMe: () => request('/users/me', { method: 'DELETE' }),

  // Chats
  getChats: () => request('/chats'),
  createChat: (name) =>
    request('/chats', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),
  getChat: (id) => request(`/chats/${id}`),
  updateChat: (id, field, value) =>
    request(`/chats/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ field, value }),
    }),
  deleteChat: (id) => request(`/chats/${id}`, { method: 'DELETE' }),
  addMember: (chatId, userId) =>
    request(`/chats/${chatId}/members`, {
      method: 'POST',
      body: JSON.stringify({ id: userId }),
    }),
  removeMember: (chatId, userId) =>
    request(`/chats/${chatId}/members/${userId}`, { method: 'DELETE' }),

  // Messages
  getMessages: (chatId, beforeID, limit = 50) => {
    const params = new URLSearchParams()
    if (beforeID) params.set('beforeID', beforeID)
    params.set('limit', limit)
    return request(`/chats/${chatId}/messages?${params}`)
  },
  postMessage: (chatId, content) =>
    request(`/chats/${chatId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    }),
  updateMessage: (chatId, messageId, content) =>
    request(`/chats/${chatId}/messages/${messageId}`, {
      method: 'PATCH',
      body: JSON.stringify({ content }),
    }),
  deleteMessage: (chatId, messageId) =>
    request(`/chats/${chatId}/messages/${messageId}`, { method: 'DELETE' }),
}