import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { permissionGroupService } from '@/services/permissionGroups'
import { folderService } from '@/services/folders'
import { userAPI } from '@/services/api'
import type { PermissionGroup, PermissionGroupFolder, PermissionGroupPermission } from '@/services/permissionGroups'
import type { Folder } from '@/services/folders'
import type { User } from '@/types'

export default function PermissionGroupManagement() {
  const { t } = useTranslation()
  const [groups, setGroups] = useState<PermissionGroup[]>([])
  const [selectedGroup, setSelectedGroup] = useState<PermissionGroup | null>(null)
  const [groupFolders, setGroupFolders] = useState<PermissionGroupFolder[]>([])
  const [groupPermissions, setGroupPermissions] = useState<PermissionGroupPermission[]>([])
  const [loading, setLoading] = useState(true)

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showAddFolderModal, setShowAddFolderModal] = useState(false)
  const [showAddPermissionModal, setShowAddPermissionModal] = useState(false)

  // Form states
  const [groupName, setGroupName] = useState('')
  const [groupDescription, setGroupDescription] = useState('')

  // Folder selection
  const [availableFolders, setAvailableFolders] = useState<Folder[]>([])
  const [selectedFolderId, setSelectedFolderId] = useState<number | null>(null)

  // User search
  const [userSearchQuery, setUserSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<User[]>([])
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [isSearching, setIsSearching] = useState(false)
  const [permissionLevel, setPermissionLevel] = useState<string>('read')

  useEffect(() => {
    loadGroups()
    loadAvailableFolders()
  }, [])

  useEffect(() => {
    if (selectedGroup) {
      loadGroupDetails(selectedGroup.id)
    }
  }, [selectedGroup])

  const loadGroups = async () => {
    try {
      setLoading(true)
      const response = await permissionGroupService.listPermissionGroups()
      setGroups(response.groups || [])
    } catch (err) {
      console.error('Failed to load permission groups:', err)
    } finally {
      setLoading(false)
    }
  }

  const loadAvailableFolders = async () => {
    try {
      const response = await folderService.listFolders()
      setAvailableFolders(response.folders || [])
    } catch (err) {
      console.error('Failed to load folders:', err)
    }
  }

  const loadGroupDetails = async (groupId: number) => {
    try {
      const [foldersRes, permsRes] = await Promise.all([
        permissionGroupService.listFolders(groupId),
        permissionGroupService.listPermissions(groupId)
      ])
      setGroupFolders(foldersRes.folders || [])
      setGroupPermissions(permsRes.permissions || [])
    } catch (err) {
      console.error('Failed to load group details:', err)
    }
  }

  const handleCreateGroup = async () => {
    if (!groupName.trim()) return

    try {
      await permissionGroupService.createPermissionGroup({
        name: groupName,
        description: groupDescription
      })
      setShowCreateModal(false)
      setGroupName('')
      setGroupDescription('')
      loadGroups()
    } catch (err) {
      console.error('Failed to create group:', err)
      alert(t('permissionGroups.createError'))
    }
  }

  const handleDeleteGroup = async (groupId: number) => {
    if (!confirm(t('permissionGroups.confirmDelete'))) return

    try {
      await permissionGroupService.deletePermissionGroup(groupId)
      if (selectedGroup?.id === groupId) {
        setSelectedGroup(null)
      }
      loadGroups()
    } catch (err) {
      console.error('Failed to delete group:', err)
      alert(t('permissionGroups.deleteError'))
    }
  }

  const handleAddFolder = async () => {
    if (!selectedGroup || !selectedFolderId) return

    try {
      await permissionGroupService.addFolder(selectedGroup.id, { folder_id: selectedFolderId })
      setShowAddFolderModal(false)
      setSelectedFolderId(null)
      loadGroupDetails(selectedGroup.id)
    } catch (err) {
      console.error('Failed to add folder:', err)
      alert(t('permissionGroups.addFolderError'))
    }
  }

  const handleRemoveFolder = async (folderId: number) => {
    if (!selectedGroup) return

    try {
      await permissionGroupService.removeFolder(selectedGroup.id, folderId)
      loadGroupDetails(selectedGroup.id)
    } catch (err) {
      console.error('Failed to remove folder:', err)
      alert(t('permissionGroups.removeFolderError'))
    }
  }

  const handleUserSearch = async (query: string) => {
    setUserSearchQuery(query)

    if (query.trim().length < 2) {
      setSearchResults([])
      return
    }

    try {
      setIsSearching(true)
      const response = await userAPI.searchUsers(query, 10)
      setSearchResults(response.data.users || [])
    } catch (err) {
      console.error('Failed to search users:', err)
      setSearchResults([])
    } finally {
      setIsSearching(false)
    }
  }

  const handleSelectUser = (user: User) => {
    setSelectedUser(user)
    setUserSearchQuery(user.username)
    setSearchResults([])
  }

  const handleGrantPermission = async () => {
    if (!selectedGroup || !selectedUser) return

    try {
      await permissionGroupService.grantPermission(selectedGroup.id, {
        user_id: selectedUser.id,
        permission: permissionLevel
      })
      setShowAddPermissionModal(false)
      setSelectedUser(null)
      setUserSearchQuery('')
      setPermissionLevel('read')
      loadGroupDetails(selectedGroup.id)
    } catch (err) {
      console.error('Failed to grant permission:', err)
      alert(t('permissionGroups.grantPermissionError'))
    }
  }

  const handleRevokePermission = async (userId: number) => {
    if (!selectedGroup) return

    try {
      await permissionGroupService.revokePermission(selectedGroup.id, userId)
      loadGroupDetails(selectedGroup.id)
    } catch (err) {
      console.error('Failed to revoke permission:', err)
      alert(t('permissionGroups.revokePermissionError'))
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">{t('permissionGroups.title')}</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg"
        >
          {t('permissionGroups.createGroup')}
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Groups List */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow p-4">
            <h2 className="text-lg font-semibold mb-4">{t('permissionGroups.groupList')}</h2>
            {loading ? (
              <div className="text-center py-4">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
              </div>
            ) : groups.length === 0 ? (
              <p className="text-gray-500 text-sm">{t('permissionGroups.noGroups')}</p>
            ) : (
              <div className="space-y-2">
                {groups.map((group) => (
                  <div
                    key={group.id}
                    onClick={() => setSelectedGroup(group)}
                    className={`p-3 rounded-lg cursor-pointer transition ${
                      selectedGroup?.id === group.id
                        ? 'bg-blue-50 border-2 border-blue-500'
                        : 'bg-gray-50 hover:bg-gray-100 border-2 border-transparent'
                    }`}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <h3 className="font-medium">{group.name}</h3>
                        {group.description && (
                          <p className="text-sm text-gray-600 mt-1 line-clamp-2">
                            {group.description}
                          </p>
                        )}
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteGroup(group.id)
                        }}
                        className="text-red-500 hover:text-red-700 ml-2"
                      >
                        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Group Details */}
        <div className="lg:col-span-2">
          {selectedGroup ? (
            <div className="space-y-6">
              {/* Folders Section */}
              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-lg font-semibold">{t('permissionGroups.foldersInGroup')}</h2>
                  <button
                    onClick={() => setShowAddFolderModal(true)}
                    className="bg-green-500 hover:bg-green-600 text-white px-3 py-1 rounded text-sm"
                  >
                    {t('permissionGroups.addFolder')}
                  </button>
                </div>
                <div className="space-y-2">
                  {groupFolders.length === 0 ? (
                    <p className="text-gray-500 text-sm">{t('permissionGroups.noFolders')}</p>
                  ) : (
                    groupFolders.map((folder) => (
                      <div key={folder.id} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                        <div className="flex-1">
                          <div className="font-medium">{folder.folder_name}</div>
                          <code className="text-sm text-gray-600">{folder.folder_path}</code>
                        </div>
                        <button
                          onClick={() => handleRemoveFolder(folder.folder_id)}
                          className="text-red-500 hover:text-red-700"
                        >
                          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* Permissions Section */}
              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-lg font-semibold">{t('permissionGroups.userPermissions')}</h2>
                  <button
                    onClick={() => setShowAddPermissionModal(true)}
                    className="bg-green-500 hover:bg-green-600 text-white px-3 py-1 rounded text-sm"
                  >
                    {t('permissionGroups.addUser')}
                  </button>
                </div>
                <div className="space-y-2">
                  {groupPermissions.length === 0 ? (
                    <p className="text-gray-500 text-sm">{t('permissionGroups.noPermissions')}</p>
                  ) : (
                    groupPermissions.map((perm) => (
                      <div key={perm.id} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                        <div>
                          <div className="font-medium">{perm.username}</div>
                          <div className="text-sm text-gray-600">{perm.email}</div>
                          <span className={`inline-block mt-1 text-xs px-2 py-1 rounded ${
                            perm.permission === 'write'
                              ? 'bg-purple-100 text-purple-800'
                              : 'bg-blue-100 text-blue-800'
                          }`}>
                            {perm.permission === 'write' ? t('permissionGroups.readWrite') : t('permissionGroups.readOnly')}
                          </span>
                        </div>
                        <button
                          onClick={() => handleRevokePermission(perm.user_id)}
                          className="text-red-500 hover:text-red-700"
                        >
                          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-white rounded-lg shadow p-12 text-center">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <p className="mt-4 text-gray-500">{t('permissionGroups.selectGroup')}</p>
            </div>
          )}
        </div>
      </div>

      {/* Create Group Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">{t('permissionGroups.createGroup')}</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('permissionGroups.groupName')}
                </label>
                <input
                  type="text"
                  value={groupName}
                  onChange={(e) => setGroupName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder={t('permissionGroups.groupNamePlaceholder')}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('permissionGroups.description')}
                </label>
                <textarea
                  value={groupDescription}
                  onChange={(e) => setGroupDescription(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows={3}
                  placeholder={t('permissionGroups.descriptionPlaceholder')}
                />
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowCreateModal(false)
                  setGroupName('')
                  setGroupDescription('')
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleCreateGroup}
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
              >
                {t('common.create')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Folder Modal */}
      {showAddFolderModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">{t('permissionGroups.addFolder')}</h2>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {t('permissionGroups.selectFolder')}
              </label>
              <select
                value={selectedFolderId || ''}
                onChange={(e) => setSelectedFolderId(Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">{t('permissionGroups.selectFolderPlaceholder')}</option>
                {availableFolders.map((folder) => (
                  <option key={folder.id} value={folder.id}>
                    {folder.name} ({folder.absolute_path})
                  </option>
                ))}
              </select>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowAddFolderModal(false)
                  setSelectedFolderId(null)
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleAddFolder}
                disabled={!selectedFolderId}
                className={`px-4 py-2 rounded-lg ${
                  selectedFolderId
                    ? 'bg-green-500 hover:bg-green-600 text-white'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                }`}
              >
                {t('common.add')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Permission Modal */}
      {showAddPermissionModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">{t('permissionGroups.addUser')}</h2>
            <div className="space-y-4">
              <div className="relative">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('permissionGroups.searchUser')}
                </label>
                <input
                  type="text"
                  value={userSearchQuery}
                  onChange={(e) => handleUserSearch(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder={t('permissionGroups.searchUserPlaceholder')}
                />

                {/* Search Results */}
                {searchResults.length > 0 && (
                  <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                    {searchResults.map((user) => (
                      <div
                        key={user.id}
                        onClick={() => handleSelectUser(user)}
                        className="px-3 py-2 hover:bg-gray-100 cursor-pointer"
                      >
                        <div className="font-medium">{user.username}</div>
                        <div className="text-sm text-gray-600">{user.email}</div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Selected User */}
                {selectedUser && (
                  <div className="mt-2 p-3 bg-blue-50 rounded-lg border border-blue-200">
                    <div className="flex justify-between items-center">
                      <div>
                        <div className="font-medium text-blue-900">{selectedUser.username}</div>
                        <div className="text-sm text-blue-700">{selectedUser.email}</div>
                      </div>
                      <button
                        onClick={() => {
                          setSelectedUser(null)
                          setUserSearchQuery('')
                        }}
                        className="text-blue-500 hover:text-blue-700"
                      >
                        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  </div>
                )}

                {isSearching && (
                  <div className="mt-2 text-sm text-gray-500 flex items-center">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500 mr-2"></div>
                    {t('permissionGroups.searching')}
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('permissionGroups.permissionLevel')}
                </label>
                <select
                  value={permissionLevel}
                  onChange={(e) => setPermissionLevel(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="read">{t('permissionGroups.readOnly')}</option>
                  <option value="write">{t('permissionGroups.readWrite')}</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowAddPermissionModal(false)
                  setSelectedUser(null)
                  setUserSearchQuery('')
                  setPermissionLevel('read')
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleGrantPermission}
                disabled={!selectedUser}
                className={`px-4 py-2 rounded-lg ${
                  selectedUser
                    ? 'bg-green-500 hover:bg-green-600 text-white'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                }`}
              >
                {t('common.add')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
