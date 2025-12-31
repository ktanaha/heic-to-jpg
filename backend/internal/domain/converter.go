package domain

// ConversionRequest はHEICからJPGへの変換リクエストを表す
type ConversionRequest struct {
	// ファイル名
	FileName string
	// HEICファイルデータ
	Data []byte
	// JPG品質（1-100、デフォルト: 85）
	Quality int
}

// ConversionResult は変換結果を表す
type ConversionResult struct {
	// 変換後のファイル名
	FileName string
	// JPGファイルデータ
	Data []byte
	// 変換前のファイルサイズ（バイト）
	OriginalSize int64
	// 変換後のファイルサイズ（バイト）
	ConvertedSize int64
}

// ImageConverter はHEIC→JPG変換のインターフェース
type ImageConverter interface {
	// Convert はHEIC画像をJPGに変換する
	Convert(req ConversionRequest) (*ConversionResult, error)
}

// Validate は変換リクエストのバリデーションを行う
func (r *ConversionRequest) Validate() error {
	if r.FileName == "" {
		return ErrInvalidFileName
	}
	if len(r.Data) == 0 {
		return ErrEmptyData
	}
	if r.Quality < 1 || r.Quality > 100 {
		return ErrInvalidQuality
	}
	return nil
}
