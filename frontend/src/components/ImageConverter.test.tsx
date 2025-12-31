import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import ImageConverter from './ImageConverter'

describe('ImageConverter', () => {
  it('コンポーネントがレンダリングされる', () => {
    render(<ImageConverter />)
    expect(screen.getByText(/HEICファイルをここにドラッグ&ドロップ/)).toBeInTheDocument()
  })

  it('品質スライダーがレンダリングされる', () => {
    render(<ImageConverter />)
    expect(screen.getByLabelText(/JPG品質:/)).toBeInTheDocument()
  })

  it('ファイル選択ボタンがレンダリングされる', () => {
    render(<ImageConverter />)
    expect(screen.getByText(/ファイルを選択/)).toBeInTheDocument()
  })
})
