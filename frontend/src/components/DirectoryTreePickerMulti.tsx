import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { folderService, DirectoryInfo, type Folder } from '@/services/folders'
import { ChevronRight, ChevronDown, Folder as FolderIcon, FolderOpen, X } from 'lucide-react'

export interface SelectedFolder {
  folderId: number
  folderName: string
  folderPath: string
  relativePath: string
  fullPath: string
}

interface DirectoryTreePickerMultiProps {
  onSelect: (selections: SelectedFolder[]) => void
  initialSelections?: SelectedFolder[]
}

interface TreeNode {
  path: string
  name: string
  isExpanded: boolean
  isLoading: boolean
  children: DirectoryInfo[]
  hasLoadedChildren: boolean
}

interface FolderTreeNode {
  folder: Folder
  treeNodes: Map<string, TreeNode>
}

export default function DirectoryTreePickerMulti({ onSelect, initialSelections = [] }: DirectoryTreePickerMultiProps) {
  const { t } = useTranslation()
  const [folders, setFolders] = useState<Folder[]>([])
  const [folderTrees, setFolderTrees] = useState<Map<number, FolderTreeNode>>(new Map())
  const [selections, setSelections] = useState<SelectedFolder[]>(initialSelections)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadFolders()
  }, [])

  // Auto-expand to selected folders when initialSelections change
  useEffect(() => {
    if (folders.length === 0 || initialSelections.length === 0) return

    const expandSelectedPaths = async () => {
      for (const selected of initialSelections) {
        const folderId = selected.folderId
        const relativePath = selected.relativePath

        if (!relativePath) {
          // Root folder selected, just expand it
          await loadDirectory(folderId, '')
        } else {
          // Need to expand all parent paths
          const pathParts = relativePath.split('/')
          let currentPath = ''

          for (const part of pathParts) {
            await loadDirectory(folderId, currentPath)
            currentPath = currentPath ? `${currentPath}/${part}` : part
          }
        }
      }
    }

    expandSelectedPaths()
  }, [folders, initialSelections])

  const loadFolders = async () => {
    try {
      const response = await folderService.listFolders()
      setFolders(response.folders || [])

      // Initialize tree nodes for each folder
      const initialTrees = new Map<number, FolderTreeNode>()
      response.folders.forEach(folder => {
        initialTrees.set(folder.id, {
          folder,
          treeNodes: new Map()
        })
      })
      setFolderTrees(initialTrees)
    } catch (error) {
      console.error('Failed to load folders:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadDirectory = async (folderId: number, relativePath: string = '') => {
    const folderTree = folderTrees.get(folderId)
    if (!folderTree) return

    const folder = folderTree.folder
    const fullPath = relativePath ? `${folder.absolute_path}/${relativePath}` : folder.absolute_path

    try {
      // Mark as loading
      setFolderTrees(prev => {
        const newMap = new Map(prev)
        const tree = newMap.get(folderId)
        if (tree) {
          const node = tree.treeNodes.get(relativePath) || {
            path: relativePath,
            name: relativePath ? relativePath.split('/').pop() || relativePath : folder.name,
            isExpanded: false,
            isLoading: true,
            children: [],
            hasLoadedChildren: false
          }
          node.isLoading = true
          tree.treeNodes.set(relativePath, node)
        }
        return newMap
      })

      const response = await folderService.browseDirectory(fullPath)

      // Update node with loaded children
      setFolderTrees(prev => {
        const newMap = new Map(prev)
        const tree = newMap.get(folderId)
        if (tree) {
          const node = tree.treeNodes.get(relativePath)
          if (node) {
            // Convert absolute paths to relative paths
            const relativeChildren = (response.directories || []).map(dir => {
              const childRelativePath = relativePath ? `${relativePath}/${dir.name}` : dir.name
              return {
                ...dir,
                path: `${folder.absolute_path}/${childRelativePath}`,
                name: dir.name
              }
            })
            node.children = relativeChildren
            node.isLoading = false
            node.hasLoadedChildren = true
            node.isExpanded = true
            tree.treeNodes.set(relativePath, node)
          }
        }
        return newMap
      })
    } catch (err: any) {
      console.error('Failed to load directory:', err)

      // Mark as not loading on error
      setFolderTrees(prev => {
        const newMap = new Map(prev)
        const tree = newMap.get(folderId)
        if (tree) {
          const node = tree.treeNodes.get(relativePath)
          if (node) {
            node.isLoading = false
            tree.treeNodes.set(relativePath, node)
          }
        }
        return newMap
      })
    }
  }

  const toggleDirectory = async (folderId: number, relativePath: string) => {
    const folderTree = folderTrees.get(folderId)
    if (!folderTree) return

    const node = folderTree.treeNodes.get(relativePath)

    if (!node) {
      // First time loading this directory - expand it
      await loadDirectory(folderId, relativePath)
    } else if (node.isExpanded) {
      // Already expanded - collapse it
      setFolderTrees(prev => {
        const newMap = new Map(prev)
        const tree = newMap.get(folderId)
        if (tree) {
          const node = tree.treeNodes.get(relativePath)
          if (node) {
            node.isExpanded = false
            tree.treeNodes.set(relativePath, node)
          }
        }
        return newMap
      })
    } else if (node.hasLoadedChildren) {
      // Collapsed but has loaded children - expand it
      setFolderTrees(prev => {
        const newMap = new Map(prev)
        const tree = newMap.get(folderId)
        if (tree) {
          const node = tree.treeNodes.get(relativePath)
          if (node) {
            node.isExpanded = true
            tree.treeNodes.set(relativePath, node)
          }
        }
        return newMap
      })
    } else {
      // Not loaded yet - load and expand it
      await loadDirectory(folderId, relativePath)
    }
  }

  const isSelected = (folderId: number, relativePath: string): boolean => {
    return selections.some(s => s.folderId === folderId && s.relativePath === relativePath)
  }

  // Check if any parent folder is selected
  const hasParentSelected = (folderId: number, relativePath: string): boolean => {
    if (!relativePath) return false // Root folder has no parent

    return selections.some(s => {
      if (s.folderId !== folderId) return false

      // Check if this selection is a parent of the current path
      if (s.relativePath === '') return true // Root folder is parent of all

      // Check if current path starts with the selected path
      return relativePath.startsWith(s.relativePath + '/')
    })
  }

  const toggleSelection = (folder: Folder, relativePath: string) => {
    const fullPath = relativePath ? `${folder.absolute_path}/${relativePath}` : folder.absolute_path

    if (isSelected(folder.id, relativePath)) {
      // Remove from selection
      const newSelections = selections.filter(
        s => !(s.folderId === folder.id && s.relativePath === relativePath)
      )
      setSelections(newSelections)
      onSelect(newSelections)
    } else {
      // Check if parent is already selected
      if (hasParentSelected(folder.id, relativePath)) {
        return // Cannot select child when parent is selected
      }

      // Remove all child selections when selecting parent
      let newSelections = selections.filter(s => {
        if (s.folderId !== folder.id) return true

        // Keep selections that are not children of current path
        if (relativePath === '') return false // Selecting root, remove all in this folder

        return !s.relativePath.startsWith(relativePath + '/')
      })

      // Add current selection
      newSelections = [
        ...newSelections,
        {
          folderId: folder.id,
          folderName: folder.name,
          folderPath: folder.absolute_path,
          relativePath,
          fullPath
        }
      ]

      setSelections(newSelections)
      onSelect(newSelections)
    }
  }

  const removeSelection = (folderId: number, relativePath: string) => {
    const newSelections = selections.filter(
      s => !(s.folderId === folderId && s.relativePath === relativePath)
    )
    setSelections(newSelections)
    onSelect(newSelections)
  }

  const clearSelections = () => {
    setSelections([])
    onSelect([])
  }

  const renderSubDirectory = (
    folder: Folder,
    dir: DirectoryInfo,
    parentPath: string,
    level: number
  ): JSX.Element => {
    const relativePath = parentPath ? `${parentPath}/${dir.name}` : dir.name
    const folderTree = folderTrees.get(folder.id)
    const node = folderTree?.treeNodes.get(relativePath)
    const selected = isSelected(folder.id, relativePath)
    const parentSelected = hasParentSelected(folder.id, relativePath)

    return (
      <div key={relativePath}>
        <div
          className={`flex items-center px-2 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer ${
            selected ? 'bg-blue-50 dark:bg-blue-900/20' : ''
          } ${parentSelected ? 'opacity-50' : ''}`}
          style={{ paddingLeft: `${level * 20 + 8}px` }}
        >
          <button
            onClick={() => toggleDirectory(folder.id, relativePath)}
            className="p-0.5 mr-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded"
          >
            {node?.isLoading ? (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
            ) : node?.isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>

          <div className="flex items-center gap-2 flex-1">
            {node?.isExpanded ? (
              <FolderOpen className="h-4 w-4 text-yellow-500" />
            ) : (
              <FolderIcon className="h-4 w-4 text-yellow-500" />
            )}
            <span className="text-sm">{dir.name}</span>
          </div>

          <input
            type="checkbox"
            checked={selected}
            disabled={parentSelected}
            onChange={() => toggleSelection(folder, relativePath)}
            onClick={(e) => e.stopPropagation()}
            className="ml-2 cursor-pointer h-4 w-4 disabled:cursor-not-allowed"
          />
        </div>

        {node?.isExpanded && node.children && node.children.length > 0 && (
          <div>
            {node.children.map(child =>
              renderSubDirectory(folder, child, relativePath, level + 1)
            )}
          </div>
        )}

        {node?.isExpanded && node.hasLoadedChildren && (!node.children || node.children.length === 0) && (
          <div
            className="text-xs text-gray-400 italic px-2 py-1"
            style={{ paddingLeft: `${(level + 1) * 20 + 8}px` }}
          >
            {t('album.noSubdirectories')}
          </div>
        )}
      </div>
    )
  }

  const renderFolder = (folder: Folder): JSX.Element => {
    const folderTree = folderTrees.get(folder.id)
    const rootNode = folderTree?.treeNodes.get('')
    const selected = isSelected(folder.id, '')

    return (
      <div key={folder.id} className="mb-3 border rounded-lg dark:border-gray-700">
        <div
          className={`flex items-center px-3 py-2.5 rounded-t-lg ${
            selected ? 'bg-blue-50 dark:bg-blue-900/20' : 'bg-gray-50 dark:bg-gray-800'
          }`}
        >
          <button
            onClick={() => toggleDirectory(folder.id, '')}
            className="p-0.5 mr-2 hover:bg-gray-200 dark:hover:bg-gray-600 rounded"
          >
            {rootNode?.isLoading ? (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-500"></div>
            ) : rootNode?.isExpanded ? (
              <ChevronDown className="h-5 w-5" />
            ) : (
              <ChevronRight className="h-5 w-5" />
            )}
          </button>

          <div className="flex items-center gap-2 flex-1">
            {rootNode?.isExpanded ? (
              <FolderOpen className="h-5 w-5 text-blue-500" />
            ) : (
              <FolderIcon className="h-5 w-5 text-blue-500" />
            )}
            <div className="flex-1">
              <div className="font-medium text-sm">{folder.name}</div>
              <div className="text-xs text-muted-foreground">{folder.absolute_path}</div>
            </div>
          </div>

          <input
            type="checkbox"
            checked={selected}
            onChange={() => toggleSelection(folder, '')}
            onClick={(e) => e.stopPropagation()}
            className="ml-2 cursor-pointer h-4 w-4"
          />
        </div>

        {rootNode?.isExpanded && rootNode.children && rootNode.children.length > 0 && (
          <div className="py-1">
            {rootNode.children.map(child =>
              renderSubDirectory(folder, child, '', 1)
            )}
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
    <div className="space-y-3">
      {/* Selection summary */}
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium">
          {t('album.selectFolders')} ({selections.length} {t('album.selected')})
        </div>
        {selections.length > 0 && (
          <button
            onClick={clearSelections}
            className="text-xs text-red-500 hover:text-red-600"
          >
            {t('album.clearSelection')}
          </button>
        )}
      </div>

      {/* Folder tree */}
      <div className="border rounded-lg p-3 max-h-96 overflow-y-auto dark:border-gray-700">
        {folders.length === 0 ? (
          <div className="text-center py-4 text-muted-foreground text-sm">
            {t('album.noFoldersAvailable')}
          </div>
        ) : (
          folders.map(folder => renderFolder(folder))
        )}
      </div>

      {/* Selected items list */}
      {selections.length > 0 && (
        <div>
          <div className="text-sm font-medium mb-2">{t('album.selectedPaths')}:</div>
          <div className="space-y-1 max-h-40 overflow-y-auto">
            {selections.map((sel, idx) => (
              <div
                key={idx}
                className="text-xs bg-blue-50 dark:bg-blue-900/20 px-3 py-2 rounded flex items-center justify-between"
              >
                <div className="flex-1 min-w-0">
                  <div className="font-medium truncate">{sel.folderName}</div>
                  <div className="text-muted-foreground truncate">
                    {sel.relativePath ? `${sel.folderPath}/${sel.relativePath}` : sel.folderPath}
                  </div>
                </div>
                <button
                  onClick={() => removeSelection(sel.folderId, sel.relativePath)}
                  className="ml-2 text-red-500 hover:text-red-600 flex-shrink-0"
                >
                  <X size={16} />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
