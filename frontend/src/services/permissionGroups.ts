import api from './api'

export interface PermissionGroup {
  id: number
  name: string
  description: string
  created_by: number
  created_at: string
  updated_at: string
}

export interface PermissionGroupFolder {
  id: number
  permission_group_id: number
  folder_id: number
  folder_name: string
  folder_path: string
  added_at: string
}

export interface PermissionGroupPermission {
  id: number
  permission_group_id: number
  user_id: number
  username: string
  email: string
  permission: string
  granted_at: string
}

export interface CreatePermissionGroupRequest {
  name: string
  description?: string
}

export interface UpdatePermissionGroupRequest {
  name?: string
  description?: string
}

export interface AddFolderRequest {
  folder_id: number
}

export interface GrantPermissionRequest {
  user_id: number
  permission: string
}

export const permissionGroupService = {
  listPermissionGroups: async (): Promise<{ groups: PermissionGroup[]; total: number }> => {
    const response = await api.get('/permission-groups')
    return response.data
  },

  getPermissionGroup: async (id: number): Promise<{ group: PermissionGroup }> => {
    const response = await api.get(`/permission-groups/${id}`)
    return response.data
  },

  createPermissionGroup: async (data: CreatePermissionGroupRequest): Promise<{ group: PermissionGroup }> => {
    const response = await api.post('/permission-groups', data)
    return response.data
  },

  updatePermissionGroup: async (id: number, data: UpdatePermissionGroupRequest): Promise<void> => {
    await api.put(`/permission-groups/${id}`, data)
  },

  deletePermissionGroup: async (id: number): Promise<void> => {
    await api.delete(`/permission-groups/${id}`)
  },

  listFolders: async (groupId: number): Promise<{ folders: PermissionGroupFolder[]; total: number }> => {
    const response = await api.get(`/permission-groups/${groupId}/folders`)
    return response.data
  },

  addFolder: async (groupId: number, data: AddFolderRequest): Promise<void> => {
    await api.post(`/permission-groups/${groupId}/folders`, data)
  },

  removeFolder: async (groupId: number, folderId: number): Promise<void> => {
    await api.delete(`/permission-groups/${groupId}/folders/${folderId}`)
  },

  listPermissions: async (groupId: number): Promise<{ permissions: PermissionGroupPermission[]; total: number }> => {
    const response = await api.get(`/permission-groups/${groupId}/permissions`)
    return response.data
  },

  grantPermission: async (groupId: number, data: GrantPermissionRequest): Promise<void> => {
    await api.post(`/permission-groups/${groupId}/permissions`, data)
  },

  revokePermission: async (groupId: number, userId: number): Promise<void> => {
    await api.delete(`/permission-groups/${groupId}/permissions/${userId}`)
  }
}
