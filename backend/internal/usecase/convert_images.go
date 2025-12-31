package usecase

import (
	"fmt"
	"sync"

	"github.com/heic-to-jpg/backend/internal/domain"
)

// ConvertImagesUsecase は画像変換のユースケース
type ConvertImagesUsecase struct {
	converter domain.ImageConverter
}

// NewConvertImagesUsecase は新しいConvertImagesUsecaseを作成する
func NewConvertImagesUsecase(converter domain.ImageConverter) *ConvertImagesUsecase {
	return &ConvertImagesUsecase{
		converter: converter,
	}
}

// BatchConversionResult はバッチ変換の結果
type BatchConversionResult struct {
	// 成功した変換結果
	Results []*domain.ConversionResult
	// 失敗した変換のエラー
	Errors map[string]error
}

// Execute は画像変換を実行する（単一ファイル）
func (u *ConvertImagesUsecase) Execute(req domain.ConversionRequest) (*domain.ConversionResult, error) {
	return u.converter.Convert(req)
}

// ExecuteBatch は複数の画像を並列で変換する
func (u *ConvertImagesUsecase) ExecuteBatch(requests []domain.ConversionRequest) *BatchConversionResult {
	result := &BatchConversionResult{
		Results: make([]*domain.ConversionResult, 0),
		Errors:  make(map[string]error),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, req := range requests {
		wg.Add(1)
		go func(r domain.ConversionRequest) {
			defer wg.Done()

			convResult, err := u.converter.Convert(r)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Errors[r.FileName] = err
			} else {
				result.Results = append(result.Results, convResult)
			}
		}(req)
	}

	wg.Wait()
	return result
}

// Validate はバッチリクエストのバリデーション
func ValidateBatchRequest(requests []domain.ConversionRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("リクエストが空です")
	}
	if len(requests) > 100 {
		return fmt.Errorf("一度に変換できるファイルは最大100個です")
	}
	return nil
}
