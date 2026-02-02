import { useState, useEffect } from 'react'
import {
  Users, UserPlus, Search, Download, Trash2, CheckCircle,
  XCircle, Key, History, Edit, AlertTriangle
} from 'lucide-react'
import { userAPI, User, UserActivityLog } from '@/services/users'
import { useAuth } from '@/contexts/AuthContext'
import Modal from '@/components/Modal'

export default function UserManagement() {
  const { user: currentUser } = useAuth()

  // State
  const [users, setUsers] = useState<User[]>([])
  const [stats, setStats] = useState({ total_users: 0, active_users: 0, admins: 0, disabled_users: 0 })
  const [selectedUsers, setSelectedUsers] = useState<Set<number>>(new Set())
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(25)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState('')
  const [loading, setLoading] = useState(false)

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showBulkDeleteModal, setShowBulkDeleteModal] = useState(false)
  const [showActivityModal, setShowActivityModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [activityLogs, setActivityLogs] = useState<UserActivityLog[]>([])

  // Form states
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    role: 'user'
  })
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  // Load users
  const loadUsers = async () => {
    try {
      setLoading(true)
      const response = await userAPI.listUsers({
        page: currentPage,
        limit: pageSize,
        search: searchQuery,
        role: roleFilter
      })
      setUsers(response.data.users || [])
      setTotal(response.data.total)
      setTotalPages(response.data.total_pages)
    } catch (error) {
      console.error('Failed to load users:', error)
      alert('Failed to load users')
    } finally {
      setLoading(false)
    }
  }

  // Load stats
  const loadStats = async () => {
    try {
      const response = await userAPI.getUserStats()
      setStats(response.data)
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }

  useEffect(() => {
    loadUsers()
    loadStats()
  }, [currentPage, pageSize, searchQuery, roleFilter])

  // Search handler with debounce
  useEffect(() => {
    const timer = setTimeout(() => {
      setCurrentPage(1)
      loadUsers()
    }, 300)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // Selection handlers
  const toggleSelectAll = () => {
    if (selectedUsers.size === users.length) {
      setSelectedUsers(new Set())
    } else {
      // Exclude current user and server_owner from selection
      setSelectedUsers(new Set(users.filter(u => u.id !== currentUser?.id && u.role !== 'server_owner').map(u => u.id)))
    }
  }

  const toggleSelect = (id: number) => {
    const newSet = new Set(selectedUsers)
    if (newSet.has(id)) {
      newSet.delete(id)
    } else {
      newSet.add(id)
    }
    setSelectedUsers(newSet)
  }

  // Create user
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    const errors: Record<string, string> = {}

    if (!formData.username) errors.username = 'Username is required'
    if (!formData.password) errors.password = 'Password is required'
    if (formData.password.length < 8) errors.password = 'Password must be at least 8 characters'
    if (formData.password !== formData.confirmPassword) errors.confirmPassword = 'Passwords do not match'

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors)
      return
    }

    try {
      await userAPI.createUser({
        username: formData.username,
        password: formData.password,
        email: formData.email,
        role: formData.role
      })
      alert('User created successfully')
      setShowCreateModal(false)
      resetForm()
      loadUsers()
      loadStats()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to create user')
    }
  }

  // Edit user
  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedUser) return

    try {
      await userAPI.updateUser(selectedUser.id, {
        email: formData.email,
        role: formData.role
      })
      alert('User updated successfully')
      setShowEditModal(false)
      resetForm()
      loadUsers()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update user')
    }
  }

  // Reset password
  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedUser) return

    const errors: Record<string, string> = {}
    if (!formData.password) errors.password = 'Password is required'
    if (formData.password.length < 8) errors.password = 'Password must be at least 8 characters'
    if (formData.password !== formData.confirmPassword) errors.confirmPassword = 'Passwords do not match'

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors)
      return
    }

    try {
      await userAPI.resetPassword(selectedUser.id, formData.password)
      alert('Password reset successfully')
      setShowPasswordModal(false)
      resetForm()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to reset password')
    }
  }

  // Delete user
  const handleDelete = async () => {
    if (!selectedUser) return

    try {
      await userAPI.deleteUser(selectedUser.id)
      alert('User deleted successfully')
      setShowDeleteModal(false)
      setSelectedUser(null)
      loadUsers()
      loadStats()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete user')
    }
  }

  // Bulk delete
  const handleBulkDelete = async () => {
    try {
      await userAPI.bulkDelete(Array.from(selectedUsers))
      alert(`${selectedUsers.size} users deleted successfully`)
      setShowBulkDeleteModal(false)
      setSelectedUsers(new Set())
      loadUsers()
      loadStats()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete users')
    }
  }

  // Bulk enable/disable
  const handleBulkToggle = async (enabled: boolean) => {
    try {
      await userAPI.bulkEnableDisable(Array.from(selectedUsers), enabled)
      alert(`${selectedUsers.size} users ${enabled ? 'enabled' : 'disabled'} successfully`)
      setSelectedUsers(new Set())
      loadUsers()
      loadStats()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update users')
    }
  }

  // Toggle single user
  const handleToggle = async (user: User) => {
    try {
      await userAPI.toggleUser(user.id, !user.enabled)
      alert(`User ${!user.enabled ? 'enabled' : 'disabled'} successfully`)
      loadUsers()
      loadStats()
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to toggle user')
    }
  }

  // View activity
  const handleViewActivity = async (user: User) => {
    setSelectedUser(user)
    setShowActivityModal(true)
    try {
      const response = await userAPI.getUserActivityLogs(user.id, 1, 20)
      setActivityLogs(response.data.logs || [])
    } catch (error) {
      console.error('Failed to load activity logs:', error)
    }
  }

  // Export users
  const handleExport = async () => {
    try {
      const response = await userAPI.exportUsers()
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'users_export.csv')
      document.body.appendChild(link)
      link.click()
      link.remove()
      alert('Users exported successfully')
    } catch (error) {
      alert('Failed to export users')
    }
  }

  // Open modals
  const openCreateModal = () => {
    resetForm()
    setShowCreateModal(true)
  }

  const openEditModal = (user: User) => {
    setSelectedUser(user)
    setFormData({
      username: user.username,
      email: user.email,
      password: '',
      confirmPassword: '',
      role: user.role
    })
    setFormErrors({})
    setShowEditModal(true)
  }

  const openPasswordModal = (user: User) => {
    setSelectedUser(user)
    setFormData({ ...formData, password: '', confirmPassword: '' })
    setFormErrors({})
    setShowPasswordModal(true)
  }

  const openDeleteModal = (user: User) => {
    setSelectedUser(user)
    setShowDeleteModal(true)
  }

  const resetForm = () => {
    setFormData({
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      role: 'user'
    })
    setFormErrors({})
    setSelectedUser(null)
  }

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold mb-2">User Management</h1>
        <p className="text-gray-600">Manage system users and permissions</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white p-6 rounded-lg border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Total Users</p>
              <p className="text-2xl font-bold">{stats.total_users}</p>
            </div>
            <Users className="text-blue-500" size={32} />
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Active Users</p>
              <p className="text-2xl font-bold">{stats.active_users}</p>
            </div>
            <CheckCircle className="text-green-500" size={32} />
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Administrators</p>
              <p className="text-2xl font-bold">{stats.admins}</p>
            </div>
            <Users className="text-purple-500" size={32} />
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Disabled Users</p>
              <p className="text-2xl font-bold">{stats.disabled_users}</p>
            </div>
            <XCircle className="text-red-500" size={32} />
          </div>
        </div>
      </div>

      {/* Toolbar */}
      <div className="bg-white p-4 rounded-lg border space-y-4">
        <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
          {/* Search and Filter */}
          <div className="flex gap-4 w-full md:w-auto">
            <div className="relative flex-1 md:w-64">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
              <input
                type="text"
                placeholder="Search users..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <select
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
              className="px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Roles</option>
              <option value="server_owner">Server Owner</option>
              <option value="admin">Administrator</option>
              <option value="user">User</option>
            </select>
          </div>

          {/* Actions */}
          <div className="flex gap-2 w-full md:w-auto">
            <button
              onClick={openCreateModal}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <UserPlus size={20} />
              Create User
            </button>
            <button
              onClick={handleExport}
              className="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              <Download size={20} />
              Export
            </button>
          </div>
        </div>

        {/* Bulk Actions */}
        {selectedUsers.size > 0 && (
          <div className="flex items-center gap-4 p-3 bg-blue-50 rounded-lg">
            <span className="font-medium">{selectedUsers.size} selected</span>
            <button
              onClick={() => handleBulkToggle(true)}
              className="px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700"
            >
              Enable
            </button>
            <button
              onClick={() => handleBulkToggle(false)}
              className="px-3 py-1 bg-yellow-600 text-white rounded hover:bg-yellow-700"
            >
              Disable
            </button>
            <button
              onClick={() => setShowBulkDeleteModal(true)}
              className="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700"
            >
              Delete
            </button>
          </div>
        )}
      </div>

      {/* Users Table */}
      <div className="bg-white rounded-lg border overflow-hidden">
        {loading ? (
          <div className="p-8 text-center">Loading...</div>
        ) : users.length === 0 ? (
          <div className="p-8 text-center text-gray-500">No users found</div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="px-4 py-3 text-left">
                      <input
                        type="checkbox"
                        checked={selectedUsers.size === users.filter(u => u.id !== currentUser?.id && u.role !== 'server_owner').length && users.filter(u => u.id !== currentUser?.id && u.role !== 'server_owner').length > 0}
                        onChange={toggleSelectAll}
                        className="rounded"
                      />
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-600">Username</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-600">Email</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-600">Role</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-600">Status</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-600">Last Login</th>
                    <th className="px-4 py-3 text-right text-sm font-semibold text-gray-600">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {users.map((user) => {
                    const isServerOwnerUser = user.role === 'server_owner'
                    const isCurrentUser = user.id === currentUser?.id
                    const canModify = !isServerOwnerUser && !isCurrentUser

                    return (
                    <tr key={user.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={selectedUsers.has(user.id)}
                          onChange={() => toggleSelect(user.id)}
                          disabled={isCurrentUser || isServerOwnerUser}
                          className="rounded"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{user.username}</span>
                          {isCurrentUser && (
                            <span className="text-xs text-gray-500">(You)</span>
                          )}
                          {isServerOwnerUser && (
                            <span className="text-xs bg-purple-100 text-purple-700 px-2 py-0.5 rounded">Protected</span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-gray-600">{user.email || '-'}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 text-xs rounded-full ${
                          user.role === 'server_owner' ? 'bg-purple-100 text-purple-700' :
                          user.role === 'admin' ? 'bg-blue-100 text-blue-700' :
                          'bg-gray-100 text-gray-700'
                        }`}>
                          {user.role === 'server_owner' ? 'Server Owner' :
                           user.role === 'admin' ? 'Admin' : 'User'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 text-xs rounded-full ${
                          user.enabled ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                        }`}>
                          {user.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">
                        {user.last_login_at ? new Date(user.last_login_at).toLocaleString() : 'Never'}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-2">
                          {canModify && (
                            <>
                              <button
                                onClick={() => openEditModal(user)}
                                className="p-1 hover:bg-gray-100 rounded"
                                title="Edit"
                              >
                                <Edit size={16} />
                              </button>
                              <button
                                onClick={() => openPasswordModal(user)}
                                className="p-1 hover:bg-gray-100 rounded"
                                title="Reset Password"
                              >
                                <Key size={16} />
                              </button>
                              <button
                                onClick={() => handleToggle(user)}
                                className="p-1 hover:bg-gray-100 rounded"
                                title={user.enabled ? 'Disable' : 'Enable'}
                              >
                                {user.enabled ? <XCircle size={16} /> : <CheckCircle size={16} />}
                              </button>
                              <button
                                onClick={() => openDeleteModal(user)}
                                className="p-1 hover:bg-gray-100 rounded text-red-600"
                                title="Delete"
                              >
                                <Trash2 size={16} />
                              </button>
                            </>
                          )}
                          {isServerOwnerUser && (
                            <span className="text-xs text-gray-500 italic">No actions available</span>
                          )}
                          <button
                            onClick={() => handleViewActivity(user)}
                            className="p-1 hover:bg-gray-100 rounded"
                            title="Activity"
                          >
                            <History size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )})}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="px-4 py-3 border-t flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-600">Show</span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value))
                    setCurrentPage(1)
                  }}
                  className="border rounded px-2 py-1"
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="text-sm text-gray-600">
                  of {total} users
                </span>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                  className="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50"
                >
                  Previous
                </button>
                <span className="px-3 py-1">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                  className="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Create User Modal */}
      <Modal isOpen={showCreateModal} onClose={() => setShowCreateModal(false)} title="Create User">
        <form onSubmit={handleCreate} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Username *</label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {formErrors.username && <p className="text-red-500 text-sm mt-1">{formErrors.username}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Role</label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="admin">Administrator</option>
              <option value="user">User</option>
            </select>
            <p className="text-xs text-gray-500 mt-1">
              Note: Server owner accounts can only be created during system initialization
            </p>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password *</label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {formErrors.password && <p className="text-red-500 text-sm mt-1">{formErrors.password}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Confirm Password *</label>
            <input
              type="password"
              value={formData.confirmPassword}
              onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {formErrors.confirmPassword && <p className="text-red-500 text-sm mt-1">{formErrors.confirmPassword}</p>}
          </div>
          <div className="flex gap-2 justify-end pt-4">
            <button
              type="button"
              onClick={() => setShowCreateModal(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Create
            </button>
          </div>
        </form>
      </Modal>

      {/* Edit User Modal */}
      <Modal isOpen={showEditModal} onClose={() => setShowEditModal(false)} title="Edit User">
        <form onSubmit={handleEdit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Username</label>
            <input
              type="text"
              value={formData.username}
              disabled
              className="w-full px-3 py-2 border rounded-lg bg-gray-50"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Role</label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="admin">Administrator</option>
              <option value="user">User</option>
            </select>
            <p className="text-xs text-gray-500 mt-1">
              Note: Server owner role cannot be assigned or modified
            </p>
          </div>
          <div className="flex gap-2 justify-end pt-4">
            <button
              type="button"
              onClick={() => setShowEditModal(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Update
            </button>
          </div>
        </form>
      </Modal>

      {/* Reset Password Modal */}
      <Modal isOpen={showPasswordModal} onClose={() => setShowPasswordModal(false)} title="Reset Password">
        <form onSubmit={handleResetPassword} className="p-6 space-y-4">
          <p className="text-sm text-gray-600">
            Reset password for <strong>{selectedUser?.username}</strong>
          </p>
          <div>
            <label className="block text-sm font-medium mb-1">New Password *</label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {formErrors.password && <p className="text-red-500 text-sm mt-1">{formErrors.password}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Confirm Password *</label>
            <input
              type="password"
              value={formData.confirmPassword}
              onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {formErrors.confirmPassword && <p className="text-red-500 text-sm mt-1">{formErrors.confirmPassword}</p>}
          </div>
          <div className="flex gap-2 justify-end pt-4">
            <button
              type="button"
              onClick={() => setShowPasswordModal(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Reset Password
            </button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal isOpen={showDeleteModal} onClose={() => setShowDeleteModal(false)} title="Delete User" size="sm">
        <div className="p-6 space-y-4">
          <div className="flex items-center gap-3 text-red-600">
            <AlertTriangle size={24} />
            <p>Are you sure you want to delete <strong>{selectedUser?.username}</strong>?</p>
          </div>
          <p className="text-sm text-gray-600">This action cannot be undone.</p>
          <div className="flex gap-2 justify-end pt-4">
            <button
              onClick={() => setShowDeleteModal(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleDelete}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
            >
              Delete
            </button>
          </div>
        </div>
      </Modal>

      {/* Bulk Delete Modal */}
      <Modal isOpen={showBulkDeleteModal} onClose={() => setShowBulkDeleteModal(false)} title="Delete Users" size="sm">
        <div className="p-6 space-y-4">
          <div className="flex items-center gap-3 text-red-600">
            <AlertTriangle size={24} />
            <p>Are you sure you want to delete {selectedUsers.size} users?</p>
          </div>
          <p className="text-sm text-gray-600">This action cannot be undone.</p>
          <div className="flex gap-2 justify-end pt-4">
            <button
              onClick={() => setShowBulkDeleteModal(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleBulkDelete}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
            >
              Delete {selectedUsers.size} Users
            </button>
          </div>
        </div>
      </Modal>

      {/* Activity Log Modal */}
      <Modal isOpen={showActivityModal} onClose={() => setShowActivityModal(false)} title="Activity Log" size="lg">
        <div className="p-6">
          <p className="text-sm text-gray-600 mb-4">
            Activity log for <strong>{selectedUser?.username}</strong>
          </p>
          {activityLogs.length === 0 ? (
            <p className="text-center text-gray-500 py-8">No activity logs found</p>
          ) : (
            <div className="space-y-3">
              {activityLogs.map((log) => (
                <div key={log.id} className="flex items-start gap-3 p-3 border rounded-lg">
                  <History size={20} className="text-gray-400 mt-1" />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium capitalize">{log.action.replace('_', ' ')}</span>
                      <span className="text-xs text-gray-500">
                        {new Date(log.created_at).toLocaleString()}
                      </span>
                    </div>
                    {log.details && (
                      <p className="text-sm text-gray-600 mt-1">{log.details}</p>
                    )}
                    {log.ip_address && (
                      <p className="text-xs text-gray-500 mt-1">IP: {log.ip_address}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}
