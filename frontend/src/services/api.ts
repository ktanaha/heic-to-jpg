import axios from 'axios'
import { logger } from './logger'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface ConversionResult {
  file_name: string
  original_size: number
  converted_size: number
  data: Uint8Array
}

export interface BatchConversionResponse {
  success_count: number
  error_count: number
  results: ConversionResult[]
  errors?: Record<string, string>
}

/**
 * 単一ファイルをHEICからJPGに変換
 */
export async function convertSingleFile(
  file: File,
  quality: number = 85
): Promise<Blob> {
  logger.logApiCall('/api/convert/single', 'POST', {
    fileName: file.name,
    quality,
  })

  const formData = new FormData()
  formData.append('file', file)
  formData.append('quality', quality.toString())

  try {
    const response = await axios.post(`${API_BASE_URL}/api/convert/single`, formData, {
      responseType: 'blob',
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })

    logger.info('変換成功', { fileName: file.name })
    return response.data
  } catch (error) {
    logger.error('変換エラー', { fileName: file.name, error })
    throw error
  }
}

/**
 * 複数ファイルをバッチ変換
 */
export async function convertBatchFiles(
  files: File[],
  quality: number = 85
): Promise<BatchConversionResponse> {
  logger.logApiCall('/api/convert/batch', 'POST', {
    fileCount: files.length,
    quality,
  })

  const formData = new FormData()
  files.forEach((file) => {
    formData.append('files', file)
  })
  formData.append('quality', quality.toString())

  try {
    const response = await axios.post<BatchConversionResponse>(
      `${API_BASE_URL}/api/convert/batch`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )

    logger.info('バッチ変換成功', {
      successCount: response.data.success_count,
      errorCount: response.data.error_count,
    })

    return response.data
  } catch (error) {
    logger.error('バッチ変換エラー', { error })
    throw error
  }
}

/**
 * Blobをダウンロード
 */
export function downloadBlob(blob: Blob, fileName: string): void {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  logger.logUserAction('ファイルダウンロード', { fileName })
}
