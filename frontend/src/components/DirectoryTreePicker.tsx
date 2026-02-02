import { useState, useEffect } from 'react'
import { folderService, DirectoryInfo } from '@/services/folders'
import { ChevronRight, ChevronDown, Folder, FolderOpen, Home, HardDrive } from 'lucide-react'

interface DirectoryTreePickerProps {
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
}

export default function DirectoryTreePicker({ onSelect, initialPath }: DirectoryTreePickerProps) {
  const [treeNodes, setTreeNodes] = useState<Map<string, TreeNode>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedPath, setSelectedPath] = useState<string>(initialPath || '/')

  // Load root directory on mount
  useEffect(() => {
    loadDirectory('/')
  }, [])

  const loadDirectory = async (path: string) => {
    try {
      setError(null)

      // Mark as loading
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path) || {
          path,
          name: path === '/' ? 'Root' : path.split('/').pop() || path,
          isExpanded: false,
          isLoading: true,
          children: [],
          hasLoadedChildren: false
        }
        node.isLoading = true
        newMap.set(path, node)
        return newMap
      })

      const response = await folderService.browseDirectory(path)

      // Update node with loaded children
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          // Handle null directories (permission denied or other errors)
          node.children = response.directories || []
          node.isLoading = false
          node.hasLoadedChildren = true
          node.isExpanded = true
          newMap.set(path, node)
        }
        return newMap
      })

      setLoading(false)

      // Show warning if no directories returned (might be permission issue)
      if (!response.directories || response.directories.length === 0) {
        console.warn(`No directories found in ${path} - might be a permission issue`)
      }
    } catch (err: any) {
      console.error('Failed to load directory:', err)
      setError(err.response?.data?.error || 'Failed to load directory')
      setLoading(false)

      // Mark as not loading on error
      setTreeNodes(prev => {
        const newMap = new Map(prev)
        const node = newMap.get(path)
        if (node) {
          node.isLoading = false
          newMap.set(path, node)
        }
        return newMap
      })
    }
  }

  const toggleDirectory = async (path: string) => {
    // Always select the path when clicking on the folder
    setSelectedPath(path)

    const node = treeNodes.get(path)

    if (!node) {
      // First time loading this directory - expand it
      await loadDirectory(path)
    } else if (node.isExpanded) {
      // Already expanded - collapse it
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
      // Collapsed but has loaded children - expand it
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
      // Not loaded yet - load and expand it
      await loadDirectory(path)
    }
  }

  const handleSelect = (path: string) => {
    setSelectedPath(path)
  }

  const handleConfirm = () => {
    onSelect(selectedPath)
  }

  const renderDirectory = (path: string, name?: string, level: number = 0): JSX.Element | null => {
    const node = treeNodes.get(path)

    const isSelected = selectedPath === path
    const isRoot = path === '/'

    // Determine display name
    let displayName: React.ReactNode
    if (isRoot) {
      displayName = (
        <span className="flex items-center gap-1">
          <Home className="h-4 w-4" />
          Root Directory
        </span>
      )
    } else if (name) {
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
              <FolderOpen className="h-4 w-4 text-blue-500" />
            ) : (
              <Folder className="h-4 w-4 text-gray-500" />
            )}
            <span className="text-sm">{displayName}</span>
          </div>

          <input
            type="radio"
            name="selected-folder"
            checked={isSelected}
            onChange={() => handleSelect(path)}
            onClick={(e) => e.stopPropagation()}
            className="ml-2 cursor-pointer"
          />
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
            No subdirectories (may be empty or no access permission)
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
          <h3 className="font-medium text-sm">Select Directory</h3>
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
        {renderDirectory('/')}
      </div>

      <div className="p-3 border-t bg-gray-50 flex justify-end">
        <button
          onClick={handleConfirm}
          className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-medium"
        >
          Confirm Selection
        </button>
      </div>
    </div>
  )
}
