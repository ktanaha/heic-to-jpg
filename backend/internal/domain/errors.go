package domain

import "errors"

var (
	// ErrInvalidFileName はファイル名が無効な場合のエラー
	ErrInvalidFileName = errors.New("ファイル名が無効です")

	// ErrEmptyData はデータが空の場合のエラー
	ErrEmptyData = errors.New("データが空です")

	// ErrInvalidQuality は品質設定が無効な場合のエラー
	ErrInvalidQuality = errors.New("品質は1-100の範囲で指定してください")

	// ErrConversionFailed は変換が失敗した場合のエラー
	ErrConversionFailed = errors.New("画像変換に失敗しました")

	// ErrUnsupportedFormat はサポートされていないフォーマットの場合のエラー
	ErrUnsupportedFormat = errors.New("サポートされていないフォーマットです")
)
