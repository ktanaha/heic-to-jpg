import React, { useState, useCallback } from 'react'
import JSZip from 'jszip'
import { FileItem, ConversionSettings } from '../types'
import { convertSingleFile, downloadBlob } from '../services/api'
import { logger } from '../services/logger'
import './ImageConverter.css'

const ImageConverter: React.FC = () => {
  const [files, setFiles] = useState<FileItem[]>([])
  const [settings, setSettings] = useState<ConversionSettings>({ quality: 85 })
  const [isDragging, setIsDragging] = useState(false)
  const [isDownloadingZip, setIsDownloadingZip] = useState(false)

  // ファイル選択
  const handleFileSelect = useCallback((selectedFiles: FileList | null) => {
    if (!selectedFiles) return

    logger.logUserAction('ファイル選択', { count: selectedFiles.length })

    const newFiles: FileItem[] = Array.from(selectedFiles).map((file) => ({
      id: Math.random().toString(36).substring(7),
      file,
      status: 'pending' as const,
    }))

    setFiles((prev) => [...prev, ...newFiles])
  }, [])

  // ドラッグ&ドロップ
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setIsDragging(false)

      logger.logUserAction('ファイルドロップ', { count: e.dataTransfer.files.length })
      handleFileSelect(e.dataTransfer.files)
    },
    [handleFileSelect]
  )

  // ファイル変換
  const handleConvert = async (fileItem: FileItem) => {
    setFiles((prev) =>
      prev.map((f) =>
        f.id === fileItem.id ? { ...f, status: 'converting' as const } : f
      )
    )

    try {
      const result = await convertSingleFile(fileItem.file, settings.quality)

      setFiles((prev) =>
        prev.map((f) =>
          f.id === fileItem.id ? { ...f, status: 'completed' as const, result } : f
        )
      )
    } catch (error) {
      setFiles((prev) =>
        prev.map((f) =>
          f.id === fileItem.id
            ? {
                ...f,
                status: 'error' as const,
                error: error instanceof Error ? error.message : '変換に失敗しました',
              }
            : f
        )
      )
    }
  }

  // すべてのファイルを変換
  const handleConvertAll = async () => {
    logger.logUserAction('すべて変換', { count: files.length })

    for (const file of files) {
      if (file.status === 'pending') {
        await handleConvert(file)
      }
    }
  }

  // ダウンロード
  const handleDownload = (fileItem: FileItem) => {
    if (!fileItem.result) return

    const newFileName = fileItem.file.name.replace(/\.heic$/i, '.jpg')
    downloadBlob(fileItem.result, newFileName)
  }

  // ファイル削除
  const handleRemove = (fileId: string) => {
    logger.logUserAction('ファイル削除', { fileId })
    setFiles((prev) => prev.filter((f) => f.id !== fileId))
  }

  // すべてのファイルを削除
  const handleRemoveAll = () => {
    logger.logUserAction('すべてのファイルを削除', { count: files.length })
    setFiles([])
  }

  // 品質変更
  const handleQualityChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const quality = parseInt(e.target.value, 10)
    setSettings({ quality })
    logger.logUserAction('品質変更', { quality })
  }

  // 変換済みファイルをZIPでダウンロード
  const handleDownloadAllAsZip = async () => {
    const completedFiles = files.filter((f) => f.status === 'completed' && f.result)

    if (completedFiles.length === 0) {
      alert('変換済みのファイルがありません。先に「すべて変換」を実行してください。')
      return
    }

    logger.logUserAction('変換済みファイルZIPダウンロード', { count: completedFiles.length })
    setIsDownloadingZip(true)

    try {
      const zip = new JSZip()

      // 変換済みファイルをZIPに追加
      for (const fileItem of completedFiles) {
        if (fileItem.result) {
          const newFileName = fileItem.file.name.replace(/\.heic$/i, '.jpg')
          zip.file(newFileName, fileItem.result)
        }
      }

      // ZIPファイルを生成
      const zipBlob = await zip.generateAsync({ type: 'blob' })

      // ダウンロード
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)
      const zipFileName = `converted_images_${timestamp}.zip`
      downloadBlob(zipBlob, zipFileName)

      logger.info('ZIP作成成功', { fileCount: completedFiles.length })
    } catch (error) {
      console.error('ZIP作成エラー:', error)
      logger.error('ZIP作成失敗', { error })
      alert('ZIPファイルの作成に失敗しました。')
    } finally {
      setIsDownloadingZip(false)
    }
  }

  return (
    <div className="image-converter">
      <div className="settings">
        <label htmlFor="quality">
          JPG品質: {settings.quality}
          <input
            id="quality"
            type="range"
            min="1"
            max="100"
            value={settings.quality}
            onChange={handleQualityChange}
          />
        </label>
      </div>

      <div
        className={`dropzone ${isDragging ? 'dragging' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <p>HEICファイルをここにドラッグ&ドロップ</p>
        <p>または</p>
        <label htmlFor="file-input" className="file-input-label">
          ファイルを選択
          <input
            id="file-input"
            type="file"
            multiple
            accept=".heic,.heif"
            onChange={(e) => handleFileSelect(e.target.files)}
          />
        </label>
      </div>

      {files.length > 0 && (
        <>
          <div className="actions">
            <button onClick={handleConvertAll} disabled={files.every((f) => f.status !== 'pending')}>
              すべて変換
            </button>
            <button
              onClick={handleDownloadAllAsZip}
              disabled={isDownloadingZip || !files.some((f) => f.status === 'completed')}
              className="download-zip-btn"
            >
              {isDownloadingZip ? 'ZIP作成中...' : '変換済みファイルをZIPダウンロード'}
            </button>
            <button onClick={handleRemoveAll} className="remove-all-btn">
              すべて削除
            </button>
          </div>

          <div className="file-list">
            {files.map((fileItem) => (
              <div key={fileItem.id} className={`file-item ${fileItem.status}`}>
                <div className="file-info">
                  <span className="file-name">{fileItem.file.name}</span>
                  <span className="file-size">
                    {(fileItem.file.size / 1024).toFixed(2)} KB
                  </span>
                </div>

                <div className="file-status">
                  {fileItem.status === 'pending' && (
                    <button onClick={() => handleConvert(fileItem)}>変換</button>
                  )}
                  {fileItem.status === 'converting' && <span>変換中...</span>}
                  {fileItem.status === 'completed' && (
                    <button onClick={() => handleDownload(fileItem)}>ダウンロード</button>
                  )}
                  {fileItem.status === 'error' && (
                    <span className="error">{fileItem.error}</span>
                  )}
                  <button onClick={() => handleRemove(fileItem.id)}>削除</button>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

export default ImageConverter
