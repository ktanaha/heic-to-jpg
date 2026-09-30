package infrastructure_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
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

// 拡張子が.HEICでも中身がJPEGのファイル（iPhoneからの転送時に自動変換されたもの）を変換できること
func TestHeicConverter_Convert_JPEGContentWithHeicExtension(t *testing.T) {
	converter := infrastructure.NewHeicConverter()

	src := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			src.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("テスト用JPEGの作成に失敗: %v", err)
	}

	result, err := converter.Convert(domain.ConversionRequest{
		FileName: "IMG_7049.HEIC",
		Data:     buf.Bytes(),
		Quality:  85,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if result.FileName != "IMG_7049.jpg" {
		t.Errorf("FileName = %q, want %q", result.FileName, "IMG_7049.jpg")
	}
	decoded, err := jpeg.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Fatalf("出力がJPEGとしてデコードできない: %v", err)
	}
	if got := decoded.Bounds(); got.Dx() != 40 || got.Dy() != 20 {
		t.Errorf("出力サイズ = %dx%d, want 40x20", got.Dx(), got.Dy())
	}
}

// TODO: 実際のHEICファイルを使った統合テストを追加
// - 正常な変換のテスト
// - 品質設定が反映されるかのテスト
// - ファイル名の拡張子変更のテスト
