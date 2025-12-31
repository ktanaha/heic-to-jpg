export interface FileItem {
  id: string
  file: File
  status: 'pending' | 'converting' | 'completed' | 'error'
  progress?: number
  result?: Blob
  error?: string
}

export interface ConversionSettings {
  quality: number
}
