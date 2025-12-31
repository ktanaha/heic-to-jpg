package infrastructure

import (
	"bytes"
	"fmt"
	"image/jpeg"
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
