import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../api/client'
import ChatList from '../components/ChatList'
import ChatWindow from '../components/ChatWindow'

export default function ChatsPage() {
  const { user } = useAuth()
  const [chats, setChats] = useState([])
  const [activeChatId, setActiveChatId] = useState(null)
  const [loading, setLoading] = useState(true)

  const loadChats = useCallback(async () => {
    try {
      const data = await api.getChats()
      setChats(data ?? [])
    } catch (err) {
      console.error('Failed to load chats', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadChats()
  }, [loadChats])

  async function handleCreateChat() {
    const name = prompt('Chat name')
    if (!name) return
    try {
      const chat = await api.createChat(name)
      setChats((prev) => [chat, ...prev])
      setActiveChatId(chat.id)
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="h-[calc(100vh-57px)] flex">
      <aside className="w-72 border-r bg-white flex flex-col">
        <div className="p-3 border-b flex items-center justify-between">
          <h2 className="font-semibold">Chats</h2>
          <button
            onClick={handleCreateChat}
            className="text-sm px-2 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700"
          >
            + New
          </button>
        </div>
        {loading ? (
          <p className="p-4 text-sm text-gray-500">Loading chats…</p>
        ) : (
          <ChatList
            chats={chats}
            activeChatId={activeChatId}
            onSelect={setActiveChatId}
            currentUserId={user?.id}
          />
        )}
      </aside>
      <section className="flex-1 flex flex-col">
        {activeChatId ? (
          <ChatWindow
            key={activeChatId}
            chatId={activeChatId}
            currentUserId={user?.id}
          />
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-400">
            Select a chat or create a new one
          </div>
        )}
      </section>
    </div>
  )
}