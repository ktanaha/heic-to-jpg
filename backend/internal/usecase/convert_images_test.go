package usecase_test

import (
	"errors"
	"testing"

	"github.com/heic-to-jpg/backend/internal/domain"
	"github.com/heic-to-jpg/backend/internal/usecase"
)

// MockConverter はテスト用のモックコンバーター
type MockConverter struct {
	shouldFail bool
}

func (m *MockConverter) Convert(req domain.ConversionRequest) (*domain.ConversionResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if m.shouldFail {
		return nil, errors.New("変換エラー")
	}

	return &domain.ConversionResult{
		FileName:      req.FileName + ".jpg",
		Data:          []byte("converted data"),
		OriginalSize:  int64(len(req.Data)),
		ConvertedSize: 100,
	}, nil
}

func TestConvertImagesUsecase_Execute(t *testing.T) {
	converter := &MockConverter{}
	uc := usecase.NewConvertImagesUsecase(converter)

	req := domain.ConversionRequest{
		FileName: "test.heic",
		Data:     []byte("test data"),
		Quality:  85,
	}

	result, err := uc.Execute(req)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result == nil {
		t.Error("結果がnilです")
	}
}

func TestConvertImagesUsecase_ExecuteBatch(t *testing.T) {
	converter := &MockConverter{}
	uc := usecase.NewConvertImagesUsecase(converter)

	requests := []domain.ConversionRequest{
		{
			FileName: "test1.heic",
			Data:     []byte("test data 1"),
			Quality:  85,
		},
		{
			FileName: "test2.heic",
			Data:     []byte("test data 2"),
			Quality:  90,
		},
	}

	result := uc.ExecuteBatch(requests)

	if len(result.Results) != 2 {
		t.Errorf("結果の数が異なります。期待: 2, 実際: %d", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("エラーが発生しています: %v", result.Errors)
	}
}

func TestValidateBatchRequest(t *testing.T) {
	tests := []struct {
		name    string
		reqs    []domain.ConversionRequest
		wantErr bool
	}{
		{
			name:    "空のリクエスト",
			reqs:    []domain.ConversionRequest{},
			wantErr: true,
		},
		{
			name: "正常なリクエスト",
			reqs: []domain.ConversionRequest{
				{FileName: "test.heic", Data: []byte("data"), Quality: 85},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := usecase.ValidateBatchRequest(tt.reqs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBatchRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
