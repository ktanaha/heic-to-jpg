package domain_test

import (
	"testing"

	"github.com/heic-to-jpg/backend/internal/domain"
)

func TestConversionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.ConversionRequest
		wantErr error
	}{
		{
			name: "正常なリクエスト",
			req: domain.ConversionRequest{
				FileName: "test.heic",
				Data:     []byte("dummy data"),
				Quality:  85,
			},
			wantErr: nil,
		},
		{
			name: "ファイル名が空",
			req: domain.ConversionRequest{
				FileName: "",
				Data:     []byte("dummy data"),
				Quality:  85,
			},
			wantErr: domain.ErrInvalidFileName,
		},
		{
			name: "データが空",
			req: domain.ConversionRequest{
				FileName: "test.heic",
				Data:     []byte{},
				Quality:  85,
			},
			wantErr: domain.ErrEmptyData,
		},
		{
			name: "品質が範囲外（小さすぎる）",
			req: domain.ConversionRequest{
				FileName: "test.heic",
				Data:     []byte("dummy data"),
				Quality:  0,
			},
			wantErr: domain.ErrInvalidQuality,
		},
		{
			name: "品質が範囲外（大きすぎる）",
			req: domain.ConversionRequest{
				FileName: "test.heic",
				Data:     []byte("dummy data"),
				Quality:  101,
			},
			wantErr: domain.ErrInvalidQuality,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
