export interface FileItem {
  id: string
  file: File
  status: 'pending' | 'converting' | 'completed' | 'error'
  progress?: number
  result?: Blob
  thumbnailUrl?: string
  rotation?: number // 回転角度（0, 90, 180, 270）
  checked?: boolean // チェックボックスの状態
  error?: string
}

export interface ConversionSettings {
  quality: number
}
