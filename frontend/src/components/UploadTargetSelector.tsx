import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { folderService, DirectoryInfo, Folder } from '@/services/folders'
import { uploadService } from '@/services/upload'
import { ChevronRight, ChevronDown, Folder as FolderIcon, FolderOpen, HardDrive, FolderPlus } from 'lucide-react'

interface UploadTargetSelectorProps {
  onSelect: (path: string) => void
  initialPath?: string
}

interface TreeNode {
  path: string
  name: string
  isExpanded: boolean
  isLoading: boolean
  children: DirectoryInfo[]
  hasLoadedChildren: boolean
  isFolder?: boolean // Is this a configured folder (from folder management)
}

export default function UploadTargetSelector({ onSelect, initialPath }: UploadTargetSelectorProps) {
  const { t } = useTranslation()
  const [treeNodes, setTreeNodes] = useState<Map<string, TreeNode>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedPath, setSelectedPath] = useState<string>(initialPath || '')
  const [folders, setFolders] = useState<Folder[]>([])
  const [showNewFolderModal, setShowNewFolderModal] = useState(false)
  const [newFolderParent, setNewFolderParent] = useState('')
  const [newFolderName, setNewFolderName] = useState('')

  // Load configured folders on mount
  useEffect(() => {
    loadConfiguredFolders()
  }, [])

  const loadConfiguredFolders = async () => {
    try {
      setLoading(true)
      const response = await folderService.listFolders()
      setFolders(response.folders || [])

      // Initialize tree nodes for each folder
      const newMap = new Map<string, TreeNode>()
      response.folders.forEach(folder => {
        newMap.set(folder.absolute_path, {
          path: folder.absolute_path,
          name: folder.name,
          isExpanded: false,
          isLoading: false,
          children: [],
          hasLoadedChildren: false,
          isFolder: true
        })
      })
      setTreeNodes(newMap)
      setLoading(false)
    } catch (err) {
      console.error('Failed to load folders:', err)
      setError('Failed to load folders')
      setLoading(false)
    }
  }

  const loadDirectory = async (path: string, folderId?: number) => {
    try {
      setError(null)

      // Mark as loading
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.isLoading = true
          newMap.set(path, node)
        }
        return newMap
      })

      const response = await uploadService.browseUploadTarget(path, folderId)

      // Update node with loaded children
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.children = response.directories || []
          node.isLoading = false
          node.hasLoadedChildren = true
          node.isExpanded = true
          newMap.set(path, node)
        }
        return newMap
      })
    } catch (err: any) {
      console.error('Failed to load directory:', err)
      setError(err.response?.data?.error || 'Failed to load directory')

      // Mark as not loading on error
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.isLoading = false
          node.children = []
          node.hasLoadedChildren = true
          node.isExpanded = true
          newMap.set(path, node)
        }
        return newMap
      })
    }
  }

  const toggleDirectory = async (path: string) => {
    setSelectedPath(path)

    const node = treeNodes.get(path)

    if (!node) {
      await loadDirectory(path)
    } else if (node.isExpanded) {
      // Collapse
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.isExpanded = false
          newMap.set(path, node)
        }
        return newMap
      })
    } else if (node.hasLoadedChildren) {
      // Expand
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.isExpanded = true
          newMap.set(path, node)
        }
        return newMap
      })
    } else {
      // Load and expand
      await loadDirectory(path)
    }
  }

  const handleSelect = (path: string) => {
    setSelectedPath(path)
  }

  const handleConfirm = () => {
    if (selectedPath) {
      onSelect(selectedPath)
    }
  }

  const handleCreateFolder = async () => {
    if (!newFolderName.trim() || !newFolderParent) return

    try {
      await uploadService.createDirectory(newFolderParent, newFolderName)
      setShowNewFolderModal(false)
      setNewFolderName('')
      // Reload the parent directory
      await loadDirectory(newFolderParent)
    } catch (err: any) {
      console.error('Failed to create folder:', err)
      alert(err.response?.data?.error || 'Failed to create folder')
    }
  }

  const openNewFolderModal = (parentPath: string) => {
    setNewFolderParent(parentPath)
    setNewFolderName('')
    setShowNewFolderModal(true)
  }

  const renderDirectory = (path: string, name?: string, level: number = 0): JSX.Element | null => {
    const node = treeNodes.get(path)
    const isSelected = selectedPath === path

    let displayName: React.ReactNode
    if (name) {
      displayName = name
    } else if (node?.name) {
      displayName = node.name
    } else {
      displayName = path.split('/').pop() || path
    }

    return (
      <div key={path}>
        <div
          className={`flex items-center px-2 py-1.5 cursor-pointer hover:bg-gray-100 rounded ${
            isSelected ? 'bg-blue-50 border-l-2 border-blue-500' : ''
          }`}
          style={{ paddingLeft: `${level * 16 + 8}px` }}
          onClick={() => toggleDirectory(path)}
        >
          <div className="p-0.5 mr-1 flex items-center">
            {node?.isLoading ? (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
            ) : node?.isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </div>

          <div className="flex items-center gap-1.5 flex-1">
            {node?.isExpanded ? (
              <FolderOpen className={`h-4 w-4 ${node?.isFolder ? 'text-blue-500' : 'text-yellow-500'}`} />
            ) : (
              <FolderIcon className={`h-4 w-4 ${node?.isFolder ? 'text-blue-500' : 'text-gray-500'}`} />
            )}
            <span className="text-sm">{displayName}</span>
            {node?.isFolder && (
              <span className="text-xs bg-blue-100 text-blue-600 px-1.5 py-0.5 rounded ml-2">Configured</span>
            )}
          </div>

          <input
            type="radio"
            name="selected-folder"
            checked={isSelected}
            onChange={() => handleSelect(path)}
            onClick={(e) => e.stopPropagation()}
            className="ml-2 cursor-pointer"
          />

          <button
            onClick={(e) => {
              e.stopPropagation()
              openNewFolderModal(path)
            }}
            className="ml-2 p-1 hover:bg-gray-200 rounded"
            title="Create subfolder"
          >
            <FolderPlus className="h-4 w-4 text-gray-600" />
          </button>
        </div>

        {/* Render children if expanded */}
        {node?.isExpanded && node.children && node.children.length > 0 && (
          <div>
            {node.children.map(child => renderDirectory(child.path, child.name, level + 1))}
          </div>
        )}

        {/* Show message if expanded but no children */}
        {node?.isExpanded && node.hasLoadedChildren && (!node.children || node.children.length === 0) && (
          <div
            className="text-xs text-gray-400 italic px-2 py-1"
            style={{ paddingLeft: `${(level + 1) * 16 + 8}px` }}
          >
            No subdirectories (empty or no access)
          </div>
        )}
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  return (
    <div className="border rounded-lg bg-white">
      <div className="p-3 border-b bg-gray-50">
        <div className="flex items-center gap-2">
          <HardDrive className="h-5 w-5 text-gray-600" />
          <h3 className="font-medium text-sm">Select Upload Target</h3>
        </div>
        {selectedPath && (
          <div className="mt-2 text-xs text-gray-600 bg-white px-2 py-1 rounded border">
            <span className="font-medium">Selected:</span> {selectedPath}
          </div>
        )}
      </div>

      {error && (
        <div className="mx-3 mt-3 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-600">
          {error}
        </div>
      )}

      <div className="max-h-96 overflow-y-auto p-2">
        {folders.length === 0 ? (
          <div className="text-center py-8 text-gray-500 text-sm">
            No folders configured. Please add folders in Folder Management first.
          </div>
        ) : (
          folders.map(folder => renderDirectory(folder.absolute_path))
        )}
      </div>

      <div className="p-3 border-t bg-gray-50 flex justify-end">
        <button
          onClick={handleConfirm}
          disabled={!selectedPath}
          className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Confirm Selection
        </button>
      </div>

      {/* New Folder Modal */}
      {showNewFolderModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">Create New Folder</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Parent Directory
                </label>
                <input
                  type="text"
                  value={newFolderParent}
                  disabled
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-600 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Folder Name
                </label>
                <input
                  type="text"
                  value={newFolderName}
                  onChange={(e) => setNewFolderName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Enter folder name"
                  autoFocus
                />
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => setShowNewFolderModal(false)}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateFolder}
                disabled={!newFolderName.trim()}
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
