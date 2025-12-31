package infrastructure_test

import (
	"testing"

	"github.com/heic-to-jpg/backend/internal/domain"
	"github.com/heic-to-jpg/backend/internal/infrastructure"
)

func TestHeicConverter_Convert_Validation(t *testing.T) {
	converter := infrastructure.NewHeicConverter()

	tests := []struct {
		name    string
		req     domain.ConversionRequest
		wantErr error
	}{
		{
			name: "ファイル名が空の場合エラー",
			req: domain.ConversionRequest{
				FileName: "",
				Data:     []byte("dummy"),
				Quality:  85,
			},
			wantErr: domain.ErrInvalidFileName,
		},
		{
			name: "データが空の場合エラー",
			req: domain.ConversionRequest{
				FileName: "test.heic",
				Data:     []byte{},
				Quality:  85,
			},
			wantErr: domain.ErrEmptyData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := converter.Convert(tt.req)
			if err != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: 実際のHEICファイルを使った統合テストを追加
// - 正常な変換のテスト
// - 品質設定が反映されるかのテスト
// - ファイル名の拡張子変更のテスト
