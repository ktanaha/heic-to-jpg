package infrastructure

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"path/filepath"
	"strings"

	"github.com/adrium/goheif"
	"github.com/heic-to-jpg/backend/internal/domain"
)

// HeicConverter はHEIC→JPG変換の実装
type HeicConverter struct{}

// NewHeicConverter は新しいHeicConverterを作成する
func NewHeicConverter() *HeicConverter {
	return &HeicConverter{}
}

// Convert はHEIC画像をJPGに変換する
func (c *HeicConverter) Convert(req domain.ConversionRequest) (*domain.ConversionResult, error) {
	// バリデーション
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// HEICファイルのデコード
	// goheif.Decode() は既にEXIF Orientationを適用済みの画像を返すため、
	// 追加の回転処理は不要
	img, err := goheif.Decode(bytes.NewReader(req.Data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrConversionFailed, err)
	}

	// デバッグ: デコード後の画像サイズをログ出力
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	log.Printf("[DEBUG] %s - デコード後の画像サイズ: 幅=%d, 高さ=%d", req.FileName, width, height)

	// 縦長の画像の場合、時計回りに90度回転
	if height > width {
		log.Printf("[DEBUG] %s - 縦長画像を検出、時計回りに90度回転します", req.FileName)
		img = c.rotate90(img)
		log.Printf("[DEBUG] %s - 回転後の画像サイズ: 幅=%d, 高さ=%d", req.FileName, img.Bounds().Dx(), img.Bounds().Dy())
	}

	// JPGエンコード
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: req.Quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrConversionFailed, err)
	}

	// 拡張子を.heicから.jpgに変更
	newFileName := c.changeExtension(req.FileName, ".jpg")

	return &domain.ConversionResult{
		FileName:      newFileName,
		Data:          buf.Bytes(),
		OriginalSize:  int64(len(req.Data)),
		ConvertedSize: int64(buf.Len()),
	}, nil
}

// changeExtension はファイル名の拡張子を変更する
func (c *HeicConverter) changeExtension(fileName, newExt string) string {
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)
	return nameWithoutExt + newExt
}

// rotate90 は画像を90度時計回りに回転
func (c *HeicConverter) rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rotated.Set(h-1-y, x, img.At(x, y))
		}
	}
	return rotated
}
