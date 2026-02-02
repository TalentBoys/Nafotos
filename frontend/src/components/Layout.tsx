import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Clock, Folder, Settings, Image, LogOut, Users, Shield, FolderCog, Globe, Upload, Link2 } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export default function Layout() {
  const { t, i18n } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuth()

  const toggleLanguage = () => {
    const newLang = i18n.language === 'en' ? 'zh' : 'en'
    i18n.changeLanguage(newLang)
    localStorage.setItem('language', newLang)
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const navItems = [
    { path: '/', label: t('nav.timeline'), icon: Clock },
    { path: '/folders', label: t('nav.folders'), icon: Folder },
    { path: '/albums', label: t('nav.albums'), icon: Image },
    { path: '/upload', label: 'Upload', icon: Upload },
    { path: '/share-management', label: t('nav.shareManagement'), icon: Link2 },
    ...(user?.role === 'admin' || user?.role === 'server_owner'
      ? [
          { path: '/folder-management', label: t('folderManagement.title'), icon: FolderCog },
          { path: '/permission-groups', label: t('permissionGroups.title'), icon: Shield },
          { path: '/users', label: t('nav.users'), icon: Users },
          { path: '/domain-config', label: t('nav.domainConfig'), icon: Globe }
        ]
      : []),
    { path: '/settings', label: t('nav.settings'), icon: Settings },
  ]

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b sticky top-0 z-10 bg-background/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">{t('app.name')}</h1>
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-600">
                {user?.username} ({user?.role})
              </span>
              <button
                onClick={toggleLanguage}
                className="px-4 py-2 rounded-lg bg-secondary hover:bg-secondary/80"
              >
                {i18n.language === 'en' ? '中文' : 'English'}
              </button>
              <button
                onClick={handleLogout}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-100 hover:bg-red-200 text-red-700"
              >
                <LogOut size={16} />
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      <div className="flex">
        <nav className="w-64 border-r min-h-[calc(100vh-73px)] p-4">
          <ul className="space-y-2">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.path
              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`flex items-center gap-3 px-4 py-2 rounded-lg transition-colors ${
                      isActive
                        ? 'bg-primary text-primary-foreground'
                        : 'hover:bg-secondary'
                    }`}
                  >
                    <Icon size={20} />
                    <span>{item.label}</span>
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>

        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
