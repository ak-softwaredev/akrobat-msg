import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white border-b px-6 py-3 flex items-center justify-between">
        <Link to="/" className="text-xl font-semibold text-indigo-600">
          Akrobat MSG
        </Link>
        <nav className="flex items-center gap-4 text-sm">
          <Link to="/" className="hover:text-indigo-600">
            Chats
          </Link>
          <Link to="/profile" className="hover:text-indigo-600">
            {user?.username ?? 'Profile'}
          </Link>
          <button
            onClick={handleLogout}
            className="px-3 py-1 rounded bg-gray-200 hover:bg-gray-300"
          >
            Logout
          </button>
        </nav>
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  )
}