import api from './api';

export interface Folder {
  id: number;
  name: string;
  absolute_path: string;
  enabled: boolean;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface CreateFolderRequest {
  name: string;
  absolute_path: string;
}

export interface UpdateFolderRequest {
  name: string;
  absolute_path?: string;
}

export interface DirectoryInfo {
  name: string;
  path: string;
  is_directory: boolean;
}

export const folderService = {
  // List all folders
  listFolders: async (): Promise<{ folders: Folder[]; total: number }> => {
    const response = await api.get('/folders');
    return response.data;
  },

  // Get a specific folder
  getFolder: async (id: number): Promise<{ folder: Folder }> => {
    const response = await api.get(`/folders/${id}`);
    return response.data;
  },

  // Create a new folder
  createFolder: async (data: CreateFolderRequest): Promise<{ folder: Folder }> => {
    const response = await api.post('/folders', data);
    return response.data;
  },

  // Update a folder
  updateFolder: async (id: number, data: UpdateFolderRequest): Promise<{ folder: Folder }> => {
    const response = await api.put(`/folders/${id}`, data);
    return response.data;
  },

  // Delete a folder
  deleteFolder: async (id: number): Promise<{ message: string }> => {
    const response = await api.delete(`/folders/${id}`);
    return response.data;
  },

  // Toggle folder enabled/disabled
  toggleFolder: async (id: number, enabled: boolean): Promise<{ message: string }> => {
    const response = await api.put(`/folders/${id}/toggle`, { enabled });
    return response.data;
  },

  // Trigger folder scan
  scanFolder: async (id: number): Promise<{ message: string }> => {
    const response = await api.post(`/folders/${id}/scan`);
    return response.data;
  },

  // List files in folder
  listFilesInFolder: async (
    id: number,
    params?: { limit?: number; offset?: number }
  ): Promise<{ files: any[]; total: number; limit: number; offset: number }> => {
    const response = await api.get(`/folders/${id}/files`, { params });
    return response.data;
  },

  // Browse directory tree
  browseDirectory: async (path: string = '/'): Promise<{ path: string; directories: DirectoryInfo[] }> => {
    const response = await api.post('/folders/browse', { path });
    return response.data;
  },
};
