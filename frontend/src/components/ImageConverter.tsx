import React, { useState, useCallback, useEffect } from 'react'
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
      checked: true,
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
      const thumbnailUrl = URL.createObjectURL(result)

      setFiles((prev) =>
        prev.map((f) =>
          f.id === fileItem.id ? { ...f, status: 'completed' as const, result, thumbnailUrl, rotation: 0 } : f
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

  // チェックボックスの変更
  const handleCheckChange = (fileId: string, checked: boolean) => {
    setFiles((prev) =>
      prev.map((f) => (f.id === fileId ? { ...f, checked } : f))
    )
  }

  // すべてをチェック
  const handleCheckAll = () => {
    logger.logUserAction('すべてチェック', { count: files.length })
    setFiles((prev) => prev.map((f) => ({ ...f, checked: true })))
  }

  // すべてのチェックを外す
  const handleUncheckAll = () => {
    logger.logUserAction('すべてのチェックを外す', { count: files.length })
    setFiles((prev) => prev.map((f) => ({ ...f, checked: false })))
  }

  // 選択したファイルを変換
  const handleConvertAll = async () => {
    const checkedFiles = files.filter((f) => f.checked && f.status === 'pending')
    logger.logUserAction('選択したファイルを変換', { count: checkedFiles.length })

    for (const file of checkedFiles) {
      await handleConvert(file)
    }
  }

  // ダウンロード
  const handleDownload = (fileItem: FileItem) => {
    if (!fileItem.result) return

    const newFileName = fileItem.file.name.replace(/\.heic$/i, '.jpg')
    downloadBlob(fileItem.result, newFileName)
  }

  // 画像を回転
  const handleRotate = async (fileItem: FileItem) => {
    if (!fileItem.result) return

    logger.logUserAction('画像回転', { fileId: fileItem.id })

    try {
      // 現在の回転角度を取得（0, 90, 180, 270）
      const currentRotation = fileItem.rotation || 0
      const newRotation = (currentRotation + 90) % 360

      // Blobを画像として読み込む
      const img = new Image()
      const originalUrl = URL.createObjectURL(fileItem.result)

      await new Promise((resolve, reject) => {
        img.onload = resolve
        img.onerror = reject
        img.src = originalUrl
      })

      // Canvasで画像を回転
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      if (!ctx) return

      // 90度または270度の場合は幅と高さを入れ替える
      if (newRotation === 90 || newRotation === 270) {
        canvas.width = img.height
        canvas.height = img.width
      } else {
        canvas.width = img.width
        canvas.height = img.height
      }

      // 回転処理
      ctx.translate(canvas.width / 2, canvas.height / 2)
      ctx.rotate((newRotation * Math.PI) / 180)
      ctx.drawImage(img, -img.width / 2, -img.height / 2)

      // Blobに変換
      const rotatedBlob = await new Promise<Blob>((resolve) => {
        canvas.toBlob((blob) => {
          if (blob) resolve(blob)
        }, 'image/jpeg', 0.85)
      })

      // 古いURLを解放
      URL.revokeObjectURL(originalUrl)
      if (fileItem.thumbnailUrl) {
        URL.revokeObjectURL(fileItem.thumbnailUrl)
      }

      // 新しいサムネイルURLを生成
      const newThumbnailUrl = URL.createObjectURL(rotatedBlob)

      // 状態を更新
      setFiles((prev) =>
        prev.map((f) =>
          f.id === fileItem.id
            ? { ...f, result: rotatedBlob, thumbnailUrl: newThumbnailUrl, rotation: newRotation }
            : f
        )
      )
    } catch (error) {
      logger.error('画像回転エラー', { error })
      console.error('画像回転エラー:', error)
    }
  }

  // ファイル削除
  const handleRemove = (fileId: string) => {
    logger.logUserAction('ファイル削除', { fileId })
    setFiles((prev) => {
      const fileToRemove = prev.find((f) => f.id === fileId)
      if (fileToRemove?.thumbnailUrl) {
        URL.revokeObjectURL(fileToRemove.thumbnailUrl)
      }
      return prev.filter((f) => f.id !== fileId)
    })
  }

  // すべてのファイルを削除
  const handleRemoveAll = () => {
    logger.logUserAction('すべてのファイルを削除', { count: files.length })
    files.forEach((file) => {
      if (file.thumbnailUrl) {
        URL.revokeObjectURL(file.thumbnailUrl)
      }
    })
    setFiles([])
  }

  // コンポーネントアンマウント時にすべてのObject URLを解放
  useEffect(() => {
    return () => {
      files.forEach((file) => {
        if (file.thumbnailUrl) {
          URL.revokeObjectURL(file.thumbnailUrl)
        }
      })
    }
  }, [files])

  // 品質変更
  const handleQualityChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const quality = parseInt(e.target.value, 10)
    setSettings({ quality })
    logger.logUserAction('品質変更', { quality })
  }

  // 選択した変換済みファイルをZIPでダウンロード
  const handleDownloadAllAsZip = async () => {
    const completedFiles = files.filter((f) => f.checked && f.status === 'completed' && f.result)

    if (completedFiles.length === 0) {
      alert('選択された変換済みファイルがありません。ファイルをチェックして変換してください。')
      return
    }

    logger.logUserAction('選択した変換済みファイルZIPダウンロード', { count: completedFiles.length })
    setIsDownloadingZip(true)

    try {
      const zip = new JSZip()

      // 選択された変換済みファイルをZIPに追加
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
            <button onClick={handleCheckAll}>
              すべてチェック
            </button>
            <button onClick={handleUncheckAll}>
              すべてのチェックを外す
            </button>
            <button
              onClick={handleConvertAll}
              disabled={!files.some((f) => f.checked && f.status === 'pending')}
            >
              選択したファイルを変換
            </button>
            <button
              onClick={handleDownloadAllAsZip}
              disabled={isDownloadingZip || !files.some((f) => f.checked && f.status === 'completed')}
              className="download-zip-btn"
            >
              {isDownloadingZip ? 'ZIP作成中...' : '選択したファイルをZIPダウンロード'}
            </button>
            <button onClick={handleRemoveAll} className="remove-all-btn">
              すべて削除
            </button>
          </div>

          <div className="file-list">
            {files.map((fileItem) => (
              <div key={fileItem.id} className={`file-item ${fileItem.status}`}>
                <div className="file-checkbox">
                  <input
                    type="checkbox"
                    checked={fileItem.checked || false}
                    onChange={(e) => handleCheckChange(fileItem.id, e.target.checked)}
                  />
                </div>
                {fileItem.thumbnailUrl && (
                  <div className="file-thumbnail">
                    <img src={fileItem.thumbnailUrl} alt={fileItem.file.name} />
                  </div>
                )}
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
                    <>
                      <button onClick={() => handleRotate(fileItem)} className="rotate-btn">
                        右回転
                      </button>
                      <button onClick={() => handleDownload(fileItem)}>ダウンロード</button>
                    </>
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
