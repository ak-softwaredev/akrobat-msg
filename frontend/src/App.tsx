import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import './App.css'

type User = { id: number; username: string; bio: string }
type Member = { chat_id: number; user_id: number; role: 'owner' | 'admin' | 'member' }
type Chat = { id: number; name: string; description: string; members: Member[]; created_by: number | null; created_at: string }
type Message = { id: number; chat_id: number; sender_id: number | null; content: string; created_at: string; updated_at: string }
type SocketEvent = { type: string; data: Message | { id: number } }
type Dialog = 'new-chat' | 'settings' | 'profile' | null

const API = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')
const TOKEN_KEY = 'akrobat.auth_token'
const THEME_KEY = 'akrobat.theme'

async function request<T>(path: string, token: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Authorization', `Bearer ${token}`)
  if (init.body) headers.set('Content-Type', 'application/json')
  const response = await fetch(`${API}${path}`, { ...init, headers })
  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const body = await response.json()
      message = body?.error?.message || body?.errors?.message || body?.error || message
    } catch { /* Some API responses have no body. */ }
    throw new Error(message)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

const initials = (name: string) => name.trim().slice(0, 2).toUpperCase() || '??'
const timeLabel = (value: string) => new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit' }).format(new Date(value))
const dayLabel = (value: string) => new Intl.DateTimeFormat(undefined, { weekday: 'long', month: 'long', day: 'numeric' }).format(new Date(value))

function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [theme, setTheme] = useState<'light' | 'dark'>(() => localStorage.getItem(THEME_KEY) === 'dark' ? 'dark' : 'light')
  const [user, setUser] = useState<User | null>(null)
  const [chats, setChats] = useState<Chat[]>([])
  const [activeChatId, setActiveChatId] = useState<number | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [memberUsers, setMemberUsers] = useState<Record<number, User>>({})
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login')
  const [busy, setBusy] = useState(false)
  const [loadingMessages, setLoadingMessages] = useState(false)
  const [hasOlder, setHasOlder] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [dialog, setDialog] = useState<Dialog>(null)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [draft, setDraft] = useState('')
  const [chatSearch, setChatSearch] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [authUsername, setAuthUsername] = useState('')
  const [authPassword, setAuthPassword] = useState('')
  const [newChatName, setNewChatName] = useState('')
  const [profileName, setProfileName] = useState('')
  const [profileBio, setProfileBio] = useState('')
  const [profilePassword, setProfilePassword] = useState('')
  const [chatName, setChatName] = useState('')
  const [chatDescription, setChatDescription] = useState('')
  const [memberId, setMemberId] = useState('')
  const socketRef = useRef<WebSocket | null>(null)
  const endRef = useRef<HTMLDivElement>(null)
  const isAtBottomRef = useRef(true)
  const activeChat = chats.find((chat) => chat.id === activeChatId) || null
  const role = activeChat?.members.find((member) => member.user_id === user?.id)?.role
  const canManage = role === 'owner' || role === 'admin'
  const filteredChats = useMemo(() => {
    const query = chatSearch.trim().toLowerCase()
    return query ? chats.filter((chat) => chat.name.toLowerCase().includes(query)) : chats
  }, [chats, chatSearch])

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    localStorage.setItem(THEME_KEY, theme)
  }, [theme])

  useEffect(() => {
    isAtBottomRef.current = true
  }, [activeChatId])

  useEffect(() => {
    if (isAtBottomRef.current) endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const themeToggle = <button className="theme-toggle" type="button" aria-pressed={theme === 'dark'} aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`} onClick={() => setTheme((current) => current === 'dark' ? 'light' : 'dark')}>
    <span aria-hidden="true">{theme === 'dark' ? '☼' : '☾'}</span><span>{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
  </button>

  useEffect(() => {
    if (!token) return
    let cancelled = false
    void Promise.all([request<User>('/api/v1/users/me', token), request<Chat[]>('/api/v1/chats', token)])
      .then(([self, chatList]) => {
        if (cancelled) return
        setUser(self); setProfileName(self.username); setProfileBio(self.bio || ''); setChats(chatList); setLoadingMessages(chatList.length > 0)
        setActiveChatId((current) => current && chatList.some((chat) => chat.id === current) ? current : chatList[0]?.id ?? null)
        setError('')
      }).catch((cause: unknown) => {
        if (cancelled) return
        localStorage.removeItem(TOKEN_KEY); setToken('')
        setError(cause instanceof Error ? cause.message : 'Could not connect to Akrobat.')
      })
    return () => { cancelled = true }
  }, [token])

  useEffect(() => {
    if (!token || !activeChatId) return
    let cancelled = false
    void request<Message[]>(`/api/v1/chats/${activeChatId}/messages?limit=50`, token)
      .then((result) => {
        if (cancelled) return
        const history = result.reverse()
        setMessages((current) => {
          const merged = new Map(history.map((message) => [message.id, message]))
          current.forEach((message) => merged.set(message.id, message))
          return [...merged.values()].sort((left, right) => left.id - right.id)
        })
        setHasOlder(result.length === 50)
        window.setTimeout(() => endRef.current?.scrollIntoView({ behavior: 'smooth' }), 0)
      }).catch((cause: unknown) => { if (!cancelled) setError(cause instanceof Error ? cause.message : 'Could not load messages.') })
      .finally(() => { if (!cancelled) setLoadingMessages(false) })
    return () => { cancelled = true }
  }, [token, activeChatId])

  useEffect(() => {
    if (!token || !activeChatId) return
    let stopped = false
    let retryTimer = 0
    let socket: WebSocket | null = null
    const connect = () => {
      if (stopped) return
      const url = new URL(`${API}/api/v1/chats/${activeChatId}/ws`, window.location.href)
      url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
      url.searchParams.set('token', token)
      socket = new WebSocket(url)
      socketRef.current = socket
      socket.onmessage = (event: MessageEvent<string>) => {
        try {
          const payload = JSON.parse(event.data) as SocketEvent
          if (payload.type === 'message.deleted') {
            const deletedId = (payload.data as { id: number }).id
            setMessages((current) => current.filter((message) => message.id !== deletedId))
            return
          }
          const message = payload.data as Message
          setMessages((current) => {
            const index = current.findIndex((item) => item.id === message.id)
            if (index < 0) return [...current, message]
            const next = [...current]; next[index] = message; return next
          })
        } catch { setError('Received an unreadable live chat event.') }
      }
      socket.onerror = () => socket?.close()
      socket.onopen = () => setNotice('')
      socket.onclose = () => {
        if (stopped) return
        setNotice('Live updates disconnected. Reconnecting…')
        retryTimer = window.setTimeout(connect, 2500)
      }
    }
    connect()
    return () => {
      stopped = true
      window.clearTimeout(retryTimer)
      socket?.close()
      if (socketRef.current === socket) socketRef.current = null
    }
  }, [token, activeChatId])

  useEffect(() => {
    if (!activeChat) return
    const ids = [...new Set(activeChat.members.map((member) => member.user_id))].filter((id) => id !== user?.id && !memberUsers[id])
    if (!ids.length) return
    let cancelled = false
    void Promise.all(ids.map(async (id) => {
      try { return await request<User>(`/api/v1/users/${id}`, token) }
      catch { return { id, username: `User ${id}`, bio: '' } }
    })).then((resolved) => {
      if (!cancelled) setMemberUsers((current) => ({ ...current, ...Object.fromEntries(resolved.map((member) => [member.id, member])) }))
    })
    return () => { cancelled = true }
  }, [activeChat, memberUsers, token, user?.id])

  const updateChat = (updated: Chat) => setChats((current) => current.map((chat) => chat.id === updated.id ? updated : chat))
  const handleAuth = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setBusy(true); setError('')
    try {
      const response = await fetch(`${API}/api/v1/auth/${authMode === 'register' ? 'register' : 'login'}`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: authUsername, password: authPassword }),
      })
      const body = await response.json()
      if (!response.ok) throw new Error(body?.error?.message || 'Could not sign in.')
      localStorage.setItem(TOKEN_KEY, body.auth_token); setToken(body.auth_token); setAuthPassword('')
      setNotice(authMode === 'register' ? 'Your account is ready. Say hello!' : 'Welcome back.')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not reach the server.') }
    finally { setBusy(false) }
  }

  const signOut = async () => {
    try { await request<void>('/api/v1/auth/logout', token, { method: 'POST' }) } catch { /* Local sign-out still works offline. */ }
    localStorage.removeItem(TOKEN_KEY); setToken(''); setUser(null); setChats([]); setMessages([]); setActiveChatId(null)
  }
  const createChat = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setError('')
    if (!newChatName.trim()) { setError('Space name cannot be blank.'); return }
    setBusy(true)
    try {
      const chat = await request<Chat>('/api/v1/chats', token, { method: 'POST', body: JSON.stringify({ name: newChatName.trim() }) })
      setChats((current) => [chat, ...current]); setMessages([]); setLoadingMessages(true); setHasOlder(false); setActiveChatId(chat.id); setNewChatName(''); setDialog(null); setSidebarOpen(false); setNotice(`Made a new corner: ${chat.name}`)
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not create chat.') }
    finally { setBusy(false) }
  }
  const sendMessage = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!activeChatId || !draft.trim()) return
    const content = draft.trim(); setDraft('')
    try {
      const message = await request<Message>(`/api/v1/chats/${activeChatId}/messages`, token, { method: 'POST', body: JSON.stringify({ content }) })
      setMessages((current) => current.some((item) => item.id === message.id) ? current : [...current, message])
    } catch (cause) { setDraft(content); setError(cause instanceof Error ? cause.message : 'Message did not send.') }
  }
  const saveMessage = async (event: FormEvent<HTMLFormElement>, id: number) => {
    event.preventDefault()
    if (!activeChatId || !draft.trim()) return
    try {
      const updated = await request<Message>(`/api/v1/chats/${activeChatId}/messages/${id}`, token, { method: 'PATCH', body: JSON.stringify({ content: draft.trim() }) })
      setMessages((current) => current.map((message) => message.id === id ? updated : message)); setEditingId(null); setDraft('')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not edit message.') }
  }
  const deleteMessage = async (id: number) => {
    if (!activeChatId || !window.confirm('Delete this message for everyone in the chat?')) return
    try {
      await request<void>(`/api/v1/chats/${activeChatId}/messages/${id}`, token, { method: 'DELETE' })
      setMessages((current) => current.filter((message) => message.id !== id))
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not delete message.') }
  }
  const loadOlder = async () => {
    if (!activeChatId || !messages.length) return
    setLoadingMessages(true)
    try {
      const older = await request<Message[]>(`/api/v1/chats/${activeChatId}/messages?beforeID=${messages[0].id}&limit=50`, token)
      setMessages((current) => [...older.reverse(), ...current]); setHasOlder(older.length === 50)
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not load earlier messages.') }
    finally { setLoadingMessages(false) }
  }
  const saveChat = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!activeChatId || !activeChat) return
    if (!chatName.trim()) { setError('Space name cannot be blank.'); return }
    setBusy(true); setError('')
    try {
      let updated = activeChat
      if (chatName.trim() !== activeChat.name) {
        updated = await request<Chat>(`/api/v1/chats/${activeChatId}`, token, { method: 'PATCH', body: JSON.stringify({ field: 'name', value: chatName.trim() }) })
        updateChat(updated)
      }
      if (chatDescription.trim() !== (activeChat.description || '')) {
        updated = await request<Chat>(`/api/v1/chats/${activeChatId}`, token, { method: 'PATCH', body: JSON.stringify({ field: 'description', value: chatDescription.trim() }) })
        updateChat(updated)
      }
      updateChat(updated); setNotice('Space details saved.'); setDialog(null)
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not update space details.') }
    finally { setBusy(false) }
  }
  const addMember = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!activeChatId || !Number(memberId)) return
    setBusy(true); setError('')
    try {
      await request<void>(`/api/v1/chats/${activeChatId}/members`, token, { method: 'POST', body: JSON.stringify({ id: Number(memberId) }) })
      updateChat(await request<Chat>(`/api/v1/chats/${activeChatId}`, token)); setMemberId(''); setNotice('Member added.')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not add member.') }
    finally { setBusy(false) }
  }
  const removeMember = async (id: number) => {
    if (!activeChatId || !window.confirm(`Remove ${memberUsers[id]?.username || `user ${id}`} from this space?`)) return
    setBusy(true); setError('')
    try {
      await request<void>(`/api/v1/chats/${activeChatId}/members/${id}`, token, { method: 'DELETE' })
      updateChat(await request<Chat>(`/api/v1/chats/${activeChatId}`, token)); setNotice('Member removed.')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not remove member.') }
    finally { setBusy(false) }
  }
  const deleteChat = async () => {
    if (!activeChatId || !window.confirm(`Delete “${activeChat?.name}” and all its messages? This cannot be undone.`)) return
    try {
      await request<void>(`/api/v1/chats/${activeChatId}`, token, { method: 'DELETE' })
      const remaining = chats.filter((chat) => chat.id !== activeChatId)
      setChats(remaining); setMessages([]); setLoadingMessages(remaining.length > 0); setHasOlder(false); setActiveChatId(remaining[0]?.id ?? null); setDialog(null); setNotice('Space deleted.')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not delete space.') }
  }
  const saveProfile = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setBusy(true); setError('')
    try {
      let updated = user
      if (profileName.trim() !== user?.username) updated = await request<User>('/api/v1/users/me', token, { method: 'PATCH', body: JSON.stringify({ field: 'username', value: profileName.trim() }) })
      if (profileBio.trim() !== (user?.bio || '')) updated = await request<User>('/api/v1/users/me', token, { method: 'PATCH', body: JSON.stringify({ field: 'bio', value: profileBio.trim() }) })
      if (profilePassword) await request<User>('/api/v1/users/me', token, { method: 'PATCH', body: JSON.stringify({ field: 'password', value: profilePassword }) })
      if (updated) setUser(updated)
      setProfilePassword(''); setNotice('Profile saved.'); setDialog(null)
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not update profile.') }
    finally { setBusy(false) }
  }
  const deleteAccount = async () => {
    if (!window.confirm('Delete your account and all chats you own? This cannot be undone.')) return
    try {
      await request<void>('/api/v1/users/me', token, { method: 'DELETE' })
      localStorage.removeItem(TOKEN_KEY); setToken(''); setUser(null); setChats([]); setMessages([]); setActiveChatId(null); setDialog(null)
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not delete account.') }
  }

  if (!token || !user) return <main className="auth-shell">
    <section className="auth-card">
      <a className="brand auth-brand" href="#home"><span className="brand-mark">a<span>!</span></span><span>akrobat<span className="brand-dot">.</span></span></a>
      <div className="auth-copy"><span className="eyebrow">YOUR LITTLE CORNER OF THE INTERNET</span><h1>{authMode === 'login' ? 'Good to see you.' : 'Come on in.'}</h1><p>{authMode === 'login' ? 'Pick up where the conversation left off.' : 'Make an account and find your people.'}</p></div>
      <form className="auth-form" onSubmit={(event) => void handleAuth(event)}>
        <label>Username<input required minLength={3} maxLength={20} autoComplete="username" value={authUsername} onChange={(event) => setAuthUsername(event.target.value)} placeholder="your-name" /></label>
        <label>Password<input required minLength={8} maxLength={64} type="password" autoComplete={authMode === 'login' ? 'current-password' : 'new-password'} value={authPassword} onChange={(event) => setAuthPassword(event.target.value)} placeholder="At least 8 characters" /></label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <button className="button button-primary auth-submit" disabled={busy}>{busy ? 'One sec…' : authMode === 'login' ? 'Log in' : 'Create account'}<span>→</span></button>
      </form>
      <p className="auth-switch">{authMode === 'login' ? 'New around here?' : 'Already have an account?'} <button type="button" className="text-button" onClick={() => { setAuthMode(authMode === 'login' ? 'register' : 'login'); setError('') }}>{authMode === 'login' ? 'Sign up' : 'Log in'}</button></p>
      <div className="auth-footnote"><span className="auth-status"><span className="live-dot" />Currently a small, local kind of place</span>{themeToggle}</div>
    </section>
    <aside className="auth-art" aria-hidden="true"><div className="sun-stamp">GOOD<br />VIBES<br /><span>ONLY</span></div><span className="art-star star-one">✳</span><span className="art-star star-two">✳</span><div className="art-note note-one">you had to be there<br /><b>well, now you are</b></div><div className="art-note note-two">a place for<br /><b>your people</b></div><div className="art-orbit orbit-one" /><div className="art-orbit orbit-two" /><div className="art-caption">MESSAGES, MINUS THE NOISE</div></aside>
  </main>

  return <main className={`app-shell ${activeChat ? '' : 'has-no-chat'}`}>
    <aside className={`sidebar ${sidebarOpen ? 'sidebar-open' : ''}`}>
      <div className="sidebar-top"><a className="brand" href="#home"><span className="brand-mark">a<span>!</span></span><span>akrobat<span className="brand-dot">.</span></span></a><button className="icon-button mobile-close" aria-label="Close chat list" onClick={() => setSidebarOpen(false)}>×</button></div>
      <div className="inbox-label"><span>YOUR SPACES</span><button className="icon-button add-chat" title="Create a chat" aria-label="Create a chat" onClick={() => { setError(''); setDialog('new-chat') }}>+</button></div>
      <label className="chat-search"><span aria-hidden="true">⌕</span><input value={chatSearch} onChange={(event) => setChatSearch(event.target.value)} placeholder="Find a space" aria-label="Find a chat" /></label>
      <div className="chat-list">{filteredChats.map((chat, index) => <button key={chat.id} className={`chat-link ${chat.id === activeChatId ? 'chat-link-active' : ''}`} onClick={() => { setMessages([]); setLoadingMessages(true); setHasOlder(false); setActiveChatId(chat.id); setSidebarOpen(false) }}><span className={`chat-symbol hue-${index % 4}`}>#</span><span className="chat-link-text"><b>{chat.name}</b><small>{chat.members.length} {chat.members.length === 1 ? 'person' : 'people'}</small></span>{chat.id === activeChatId && <span className="active-indicator" />}</button>)}
        {!chats.length && <div className="empty-spaces"><span>☁</span><p>No spaces yet.<br />Start one for your people.</p><button className="button button-small" onClick={() => setDialog('new-chat')}>Make a space <span>+</span></button></div>}{!!chats.length && !filteredChats.length && <p className="no-results">No spaces match that.</p>}
      </div>
      <div className="sidebar-bottom"><div className="user-card"><span className="avatar avatar-me">{initials(user.username)}</span><span className="user-card-copy"><b>{user.username}</b><small>here for a chat</small></span><button className="icon-button user-menu" title="Profile settings" aria-label="Profile settings" onClick={() => { setProfileName(user.username); setProfileBio(user.bio || ''); setError(''); setDialog('profile') }}>···</button></div><div className="sidebar-actions">{themeToggle}<button className="signout-button" onClick={() => void signOut()}><span>↪</span> Log out</button></div></div>
    </aside>
    {sidebarOpen && <button className="sidebar-backdrop" aria-label="Close navigation" onClick={() => setSidebarOpen(false)} />}

    <section className="conversation">
      {activeChat ? <>
        <header className="channel-header"><button className="icon-button mobile-menu" aria-label="Open chat list" onClick={() => setSidebarOpen(true)}>☰</button><span className="header-hash">#</span><div className="channel-heading"><h1>{activeChat.name}</h1><p>{activeChat.description || 'A little corner for your people.'}</p></div><div className="header-actions"><div className="member-stack" aria-label={`${activeChat.members.length} members`}>{activeChat.members.slice(0, 3).map((member, index) => { const name = member.user_id === user.id ? user.username : memberUsers[member.user_id]?.username || `User ${member.user_id}`; return <span key={member.user_id} className={`avatar mini-avatar hue-${index % 4}`} title={name}>{initials(name)}</span> })}</div><span className="member-count">{activeChat.members.length}</span><button className="icon-button channel-settings" title="Chat settings and members" aria-label="Chat settings and members" onClick={() => { setError(''); setChatName(activeChat.name); setChatDescription(activeChat.description || ''); setDialog('settings') }}>···</button></div></header>
        {notice && <div className="notice-bar"><span className="live-dot" />{notice}<button aria-label="Dismiss notice" onClick={() => setNotice('')}>×</button></div>}{error && <div className="error-bar" role="alert"><span>!</span>{error}<button aria-label="Dismiss error" onClick={() => setError('')}>×</button></div>}
        <div className="messages-scroll" onScroll={(event) => { const element = event.currentTarget; isAtBottomRef.current = element.scrollHeight - element.scrollTop - element.clientHeight <= 40 }}><div className="welcome-block"><span className="welcome-icon">#</span><span className="eyebrow">YOU MADE IT</span><h2>Welcome to {activeChat.name}</h2><p>This is the beginning of this little space.</p></div>
          {hasOlder && <button className="load-older" disabled={loadingMessages} onClick={() => void loadOlder()}>{loadingMessages ? 'Loading…' : '↑  Load earlier messages'}</button>}
          {messages.map((message, index) => { const sender = message.sender_id === user.id ? user : message.sender_id ? memberUsers[message.sender_id] : undefined; const previous = messages[index - 1]; const showDay = !previous || new Date(previous.created_at).toDateString() !== new Date(message.created_at).toDateString(); const grouped = previous?.sender_id === message.sender_id && !showDay && new Date(message.created_at).getTime() - new Date(previous.created_at).getTime() < 300000
            return <div key={message.id}>{showDay && <div className="day-divider"><span>{dayLabel(message.created_at)}</span></div>}<article className={`message-row ${grouped ? 'message-grouped' : ''}`}>
              {!grouped && <span className={`avatar message-avatar hue-${(message.sender_id || 0) % 4}`}>{initials(sender?.username || (message.sender_id ? `User ${message.sender_id}` : '??'))}</span>}
              <div className="message-body">{!grouped && <div className="message-meta"><b>{sender?.username || (message.sender_id ? `User ${message.sender_id}` : 'Former member')}</b><time>{timeLabel(message.created_at)}</time></div>}{editingId === message.id ? <form className="edit-message-form" onSubmit={(event) => void saveMessage(event, message.id)}><input autoFocus value={draft} onChange={(event) => setDraft(event.target.value)} maxLength={4000} /><button className="button button-small">Save</button><button className="text-button" type="button" onClick={() => { setEditingId(null); setDraft('') }}>Cancel</button></form> : <p className="message-content">{message.content}{message.updated_at !== message.created_at && <span className="edited-mark"> (edited)</span>}</p>}</div>
              {message.sender_id === user.id && editingId !== message.id && <div className="message-actions"><button title="Edit message" aria-label="Edit message" onClick={() => { setEditingId(message.id); setDraft(message.content) }}>✎</button><button title="Delete message" aria-label="Delete message" onClick={() => void deleteMessage(message.id)}>×</button></div>}
            </article></div>})}
          {!messages.length && !loadingMessages && <div className="first-message"><span>✳</span><p>It’s quiet in here. Make the first move.</p></div>}<div ref={endRef} />
        </div>
        <form className="composer" onSubmit={(event) => void sendMessage(event)}><button type="button" className="composer-plus" title="Create a new chat" aria-label="Create a new chat" onClick={() => setDialog('new-chat')}>+</button><input value={draft} onChange={(event) => setDraft(event.target.value)} placeholder={`Message #${activeChat.name}`} maxLength={4000} aria-label="Message content" /><span className="composer-hint">↵</span><button className="send-button" disabled={!draft.trim()} aria-label="Send message" title="Send message">↑</button></form>
      </> : <div className="no-chat"><span className="no-chat-art">✳</span><span className="eyebrow">A LITTLE SPACE TOGETHER</span><h1>Every good chat<br />starts somewhere.</h1><p>Make a space and bring your people in.</p><button className="button button-primary" onClick={() => setDialog('new-chat')}>Create your first space <span>→</span></button></div>}
    </section>

    {activeChat && <aside className="details-panel"><div className="details-cover"><span className="cover-spark">✳</span><span className="cover-label">YOUR SPACE</span></div><div className="details-content"><span className="eyebrow">THE LOWDOWN</span><h2>{activeChat.name}</h2><p className="details-description">{activeChat.description || 'No description yet. Give this space a little personality.'}</p><div className="detail-divider" /><div className="detail-title"><h3>Good company</h3><span>{activeChat.members.length}</span></div>
      <div className="member-list">{activeChat.members.map((member, index) => { const name = member.user_id === user.id ? user.username : memberUsers[member.user_id]?.username || `User ${member.user_id}`; return <div className="member-row" key={member.user_id}><span className={`avatar member-avatar hue-${index % 4}`}>{initials(name)}</span><span className="member-copy"><b>{name}</b><small>{member.role === 'owner' ? 'space maker' : member.role}</small></span>{canManage && member.user_id !== user.id && <button className="icon-button remove-member" title={`Remove ${name}`} aria-label={`Remove ${name}`} onClick={() => void removeMember(member.user_id)}>×</button>}</div>})}</div>
      {canManage && <button className="invite-button" onClick={() => { setError(''); setChatName(activeChat.name); setChatDescription(activeChat.description || ''); setDialog('settings') }}><span>+</span> Add someone</button>}<div className="detail-divider" /><div className="space-meta"><span>Made</span><b>{new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', year: 'numeric' }).format(new Date(activeChat.created_at))}</b></div>{canManage && <button className="manage-button" onClick={() => { setChatName(activeChat.name); setChatDescription(activeChat.description || ''); setDialog('settings') }}>Edit this space <span>↗</span></button>}</div><div className="details-footer"><span className="live-dot" />Made for hanging out</div></aside>}

    {dialog && <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setDialog(null) }}><section className="modal" role="dialog" aria-modal="true" aria-labelledby="dialog-title"><button className="icon-button modal-close" aria-label="Close dialog" onClick={() => setDialog(null)}>×</button>
      {dialog === 'settings' && activeChat ? <><span className="eyebrow">A SPACE FOR YOUR PEOPLE</span><h2 id="dialog-title">Space settings</h2>{canManage ? <form className="modal-form" onSubmit={(event) => void saveChat(event)}><label>Space name<input value={chatName} onChange={(event) => setChatName(event.target.value)} required maxLength={255} /></label><label>A short description<textarea value={chatDescription} onChange={(event) => setChatDescription(event.target.value)} rows={3} maxLength={500} placeholder="What brings this crew together?" /></label><button className="button button-primary" disabled={busy}>Save changes <span>→</span></button></form> : <p className="modal-note">Only an owner or admin can change this space.</p>}
        {canManage && <form className="invite-form" onSubmit={(event) => void addMember(event)}><div><b>Bring someone in</b><small>Use their account ID for now.</small></div><div className="invite-controls"><input type="number" min="1" value={memberId} onChange={(event) => setMemberId(event.target.value)} placeholder="Account ID" aria-label="Member account ID" required /><button className="button button-small" disabled={busy}>Add</button></div></form>}{error && <p className="form-error" role="alert">{error}</p>}{role === 'owner' && <button className="danger-button" onClick={() => void deleteChat()}>Delete this space</button>}</>
        : dialog === 'new-chat' ? <><span className="eyebrow">START SOMETHING GOOD</span><h2 id="dialog-title">Make a new space</h2><p className="modal-note">Give your group a home. You can invite people once it’s made.</p><form className="modal-form" onSubmit={(event) => void createChat(event)}><label>Space name<input autoFocus value={newChatName} onChange={(event) => setNewChatName(event.target.value)} placeholder="e.g. sunday-project" required maxLength={255} /></label>{error && <p className="form-error" role="alert">{error}</p>}<button className="button button-primary" disabled={busy}>Create space <span>→</span></button></form></>
        : <><span className="eyebrow">THE PERSON BEHIND THE MESSAGES</span><h2 id="dialog-title">Your profile</h2><form className="modal-form" onSubmit={(event) => void saveProfile(event)}><label>Username<input autoFocus value={profileName} onChange={(event) => setProfileName(event.target.value)} required minLength={3} maxLength={20} /></label><label>A little about you<textarea value={profileBio} onChange={(event) => setProfileBio(event.target.value)} rows={3} maxLength={500} placeholder="One line is plenty." /></label><label>Change password <span className="label-hint">Leave blank to keep the current one</span><input type="password" value={profilePassword} onChange={(event) => setProfilePassword(event.target.value)} minLength={profilePassword ? 8 : undefined} maxLength={64} autoComplete="new-password" /></label>{error && <p className="form-error" role="alert">{error}</p>}<button className="button button-primary" disabled={busy}>Save profile <span>→</span></button></form><button className="danger-button" onClick={() => void deleteAccount()}>Delete my account</button></>}
    </section></div>}
  </main>
}

export default App
