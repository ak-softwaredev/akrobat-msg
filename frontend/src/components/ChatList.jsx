export default function ChatList({ chats, activeChatId, onSelect }) {
  if (!chats.length) {
    return <p className="p-4 text-sm text-gray-500">No chats yet.</p>
  }

  return (
    <ul className="overflow-y-auto flex-1">
      {chats.map((chat) => (
        <li key={chat.id}>
          <button
            onClick={() => onSelect(chat.id)}
            className={`w-full text-left px-4 py-3 border-b hover:bg-gray-50 ${
              activeChatId === chat.id ? 'bg-indigo-50' : ''
            }`}
          >
            <div className="font-medium truncate">{chat.name}</div>
            {chat.description && (
              <div className="text-xs text-gray-500 truncate">
                {chat.description}
              </div>
            )}
          </button>
        </li>
      ))}
    </ul>
  )
}