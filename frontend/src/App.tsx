import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'

// Lazy load all page components
const Login = lazy(() => import('./pages/Login'))
const Timeline = lazy(() => import('./pages/Timeline'))
const Folders = lazy(() => import('./pages/Folders'))
const FileDetail = lazy(() => import('./pages/FileDetail'))
const Albums = lazy(() => import('./pages/Albums'))
const AlbumDetail = lazy(() => import('./pages/AlbumDetail'))
const Settings = lazy(() => import('./pages/Settings'))
const UserManagement = lazy(() => import('./pages/UserManagement'))
const FolderManagement = lazy(() => import('./pages/FolderManagement'))
const PermissionGroupManagement = lazy(() => import('./pages/PermissionGroupManagement'))
const DomainConfig = lazy(() => import('./pages/DomainConfig'))
const Upload = lazy(() => import('./pages/Upload'))
const ShareManagement = lazy(() => import('./pages/ShareManagement'))
const PublicShare = lazy(() => import('./pages/PublicShare'))

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Suspense fallback={
          <div className="flex items-center justify-center min-h-screen">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
              <p className="mt-4 text-gray-600">Loading...</p>
            </div>
          </div>
        }>
          <Routes>
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/s/:id" element={<PublicShare />} />

            {/* Fullscreen file detail page (no layout) */}
            <Route path="/file/:id" element={
              <ProtectedRoute>
                <FileDetail />
              </ProtectedRoute>
            } />

            {/* Protected routes with layout */}
            <Route path="/" element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }>
              <Route index element={<Timeline />} />
              <Route path="folders" element={<Folders />} />
              <Route path="albums" element={<Albums />} />
              <Route path="albums/:id" element={<AlbumDetail />} />
              <Route path="upload" element={<Upload />} />
              <Route path="folder-management" element={<FolderManagement />} />
              <Route path="permission-groups" element={<PermissionGroupManagement />} />
              <Route path="users" element={<UserManagement />} />
              <Route path="domain-config" element={<DomainConfig />} />
              <Route path="share-management" element={<ShareManagement />} />
              <Route path="settings" element={<Settings />} />
            </Route>
          </Routes>
        </Suspense>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
