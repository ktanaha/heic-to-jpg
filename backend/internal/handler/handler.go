package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/heic-to-jpg/backend/internal/domain"
	"github.com/heic-to-jpg/backend/internal/usecase"
	vibelogger "github.com/ktanaha/vibe-coding-logger/pkg/logger"
)

const (
	defaultQuality = 85
	maxFileSize    = 50 * 1024 * 1024 // 50MB
)

// ConvertHandler は画像変換のHTTPハンドラー
type ConvertHandler struct {
	usecase *usecase.ConvertImagesUsecase
	logger  vibelogger.Logger
}

// NewConvertHandler は新しいConvertHandlerを作成する
func NewConvertHandler(uc *usecase.ConvertImagesUsecase, logger vibelogger.Logger) *ConvertHandler {
	return &ConvertHandler{
		usecase: uc,
		logger:  logger,
	}
}

// ConvertResponse は変換APIのレスポンス
type ConvertResponse struct {
	FileName      string `json:"file_name"`
	OriginalSize  int64  `json:"original_size"`
	ConvertedSize int64  `json:"converted_size"`
	Data          string `json:"data"` // Base64エンコードされたデータ
}

// ConvertSingle は単一ファイルの変換を処理する
func (h *ConvertHandler) ConvertSingle(c *gin.Context) {
	log.Println("[ConvertSingle] 変換リクエスト開始")
	tracker := h.logger.StartOperation("単一ファイル変換", map[string]interface{}{})

	// ファイルを取得
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("[ConvertSingle] ファイル取得エラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "ファイルの取得に失敗")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルが見つかりません"})
		return
	}
	defer file.Close()
	log.Printf("[ConvertSingle] ファイル取得成功: %s (サイズ: %d bytes)", header.Filename, header.Size)

	// ファイルサイズチェック
	if header.Size > maxFileSize {
		log.Printf("[ConvertSingle] ファイルサイズ超過: %d bytes", header.Size)
		h.logger.ErrorOperation(tracker, fmt.Errorf("ファイルサイズが大きすぎます"), "サイズ制限超過")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルサイズは50MB以下にしてください"})
		return
	}

	// 品質パラメータを取得
	quality := defaultQuality
	if qStr := c.DefaultPostForm("quality", ""); qStr != "" {
		if q, err := strconv.Atoi(qStr); err == nil && q >= 1 && q <= 100 {
			quality = q
		}
	}
	log.Printf("[ConvertSingle] 品質設定: %d", quality)

	// ファイルデータを読み込む
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[ConvertSingle] ファイル読み込みエラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "ファイルの読み込みに失敗")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルの読み込みに失敗しました"})
		return
	}
	log.Printf("[ConvertSingle] ファイル読み込み完了: %d bytes", len(data))

	// 変換リクエストを作成
	req := domain.ConversionRequest{
		FileName: header.Filename,
		Data:     data,
		Quality:  quality,
	}

	// 変換を実行
	log.Println("[ConvertSingle] 変換開始")
	result, err := h.usecase.Execute(req)
	if err != nil {
		log.Printf("[ConvertSingle] 変換エラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "変換に失敗")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("変換に失敗しました: %v", err)})
		return
	}
	log.Printf("[ConvertSingle] 変換成功: %s (%d bytes → %d bytes)", result.FileName, result.OriginalSize, result.ConvertedSize)

	h.logger.CompleteOperation(tracker, map[string]interface{}{
		"original_size":  result.OriginalSize,
		"converted_size": result.ConvertedSize,
	})

	// 変換後のファイルを返す
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.FileName))
	c.Data(http.StatusOK, "image/jpeg", result.Data)
	log.Println("[ConvertSingle] レスポンス送信完了")
}

// ConvertBatch は複数ファイルの変換を処理する
func (h *ConvertHandler) ConvertBatch(c *gin.Context) {
	tracker := h.logger.StartOperation("バッチ変換", map[string]interface{}{})

	// マルチパートフォームを解析
	form, err := c.MultipartForm()
	if err != nil {
		h.logger.ErrorOperation(tracker, err, "フォームの解析に失敗")
		c.JSON(http.StatusBadRequest, gin.H{"error": "フォームの解析に失敗しました"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		h.logger.ErrorOperation(tracker, fmt.Errorf("ファイルがありません"), "ファイル不足")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルが見つかりません"})
		return
	}

	// 品質パラメータを取得
	quality := defaultQuality
	if qStr := c.DefaultPostForm("quality", ""); qStr != "" {
		if q, err := strconv.Atoi(qStr); err == nil && q >= 1 && q <= 100 {
			quality = q
		}
	}

	// 変換リクエストを作成
	requests := make([]domain.ConversionRequest, 0, len(files))
	for _, fileHeader := range files {
		// ファイルを開く
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		// データを読み込む
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		requests = append(requests, domain.ConversionRequest{
			FileName: fileHeader.Filename,
			Data:     data,
			Quality:  quality,
		})
	}

	// バリデーション
	if err := usecase.ValidateBatchRequest(requests); err != nil {
		h.logger.ErrorOperation(tracker, err, "バリデーションエラー")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// バッチ変換を実行
	result := h.usecase.ExecuteBatch(requests)

	h.logger.CompleteOperation(tracker, map[string]interface{}{
		"success_count": len(result.Results),
		"error_count":   len(result.Errors),
	})

	// レスポンスを作成
	response := gin.H{
		"success_count": len(result.Results),
		"error_count":   len(result.Errors),
		"results":       convertResultsToResponse(result.Results),
	}

	if len(result.Errors) > 0 {
		response["errors"] = result.Errors
	}

	c.JSON(http.StatusOK, response)
}

// convertResultsToResponse は変換結果をレスポンス形式に変換する
func convertResultsToResponse(results []*domain.ConversionResult) []gin.H {
	responses := make([]gin.H, len(results))
	for i, r := range results {
		responses[i] = gin.H{
			"file_name":      r.FileName,
			"original_size":  r.OriginalSize,
			"converted_size": r.ConvertedSize,
			"data":           r.Data, // フロントエンドでBase64エンコードする
		}
	}
	return responses
}

// isHEICFile はファイルがHEIC形式かどうかを判定する
func isHEICFile(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, ".heic") || strings.HasSuffix(lower, ".heif")
}

// ConvertBatchZip は複数ファイルを変換してZIPファイルとして返す
func (h *ConvertHandler) ConvertBatchZip(c *gin.Context) {
	log.Println("[ConvertBatchZip] バッチ変換+ZIP圧縮リクエスト開始")
	tracker := h.logger.StartOperation("バッチ変換+ZIP", map[string]interface{}{})

	// マルチパートフォームを解析
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("[ConvertBatchZip] フォーム解析エラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "フォームの解析に失敗")
		c.JSON(http.StatusBadRequest, gin.H{"error": "フォームの解析に失敗しました"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		log.Println("[ConvertBatchZip] ファイルが見つかりません")
		h.logger.ErrorOperation(tracker, fmt.Errorf("ファイルがありません"), "ファイル不足")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルが見つかりません"})
		return
	}
	log.Printf("[ConvertBatchZip] %d個のファイルを受信", len(files))

	// 品質パラメータを取得
	quality := defaultQuality
	if qStr := c.DefaultPostForm("quality", ""); qStr != "" {
		if q, err := strconv.Atoi(qStr); err == nil && q >= 1 && q <= 100 {
			quality = q
		}
	}
	log.Printf("[ConvertBatchZip] 品質設定: %d", quality)

	// 変換リクエストを作成
	requests := make([]domain.ConversionRequest, 0, len(files))
	for _, fileHeader := range files {
		// ファイルを開く
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("[ConvertBatchZip] ファイルオープンエラー: %s - %v", fileHeader.Filename, err)
			continue
		}

		// データを読み込む
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			log.Printf("[ConvertBatchZip] ファイル読み込みエラー: %s - %v", fileHeader.Filename, err)
			continue
		}

		requests = append(requests, domain.ConversionRequest{
			FileName: fileHeader.Filename,
			Data:     data,
			Quality:  quality,
		})
	}
	log.Printf("[ConvertBatchZip] %d個のファイルを変換キューに追加", len(requests))

	// バリデーション
	if err := usecase.ValidateBatchRequest(requests); err != nil {
		log.Printf("[ConvertBatchZip] バリデーションエラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "バリデーションエラー")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// バッチ変換を実行
	log.Println("[ConvertBatchZip] バッチ変換開始")
	result := h.usecase.ExecuteBatch(requests)
	log.Printf("[ConvertBatchZip] バッチ変換完了 - 成功: %d, 失敗: %d", len(result.Results), len(result.Errors))

	if len(result.Results) == 0 {
		log.Println("[ConvertBatchZip] 変換成功したファイルがありません")
		h.logger.ErrorOperation(tracker, fmt.Errorf("すべてのファイルの変換に失敗"), "全ファイル変換失敗")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "すべてのファイルの変換に失敗しました",
			"errors": result.Errors,
		})
		return
	}

	// ZIPファイルを作成
	log.Println("[ConvertBatchZip] ZIP作成開始")
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, convResult := range result.Results {
		writer, err := zipWriter.Create(convResult.FileName)
		if err != nil {
			log.Printf("[ConvertBatchZip] ZIPエントリ作成エラー: %s - %v", convResult.FileName, err)
			continue
		}

		_, err = writer.Write(convResult.Data)
		if err != nil {
			log.Printf("[ConvertBatchZip] ZIPファイル書き込みエラー: %s - %v", convResult.FileName, err)
			continue
		}
		log.Printf("[ConvertBatchZip] ZIPに追加: %s (%d bytes)", convResult.FileName, len(convResult.Data))
	}

	if err := zipWriter.Close(); err != nil {
		log.Printf("[ConvertBatchZip] ZIP終了処理エラー: %v", err)
		h.logger.ErrorOperation(tracker, err, "ZIP作成に失敗")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZIP作成に失敗しました"})
		return
	}
	log.Printf("[ConvertBatchZip] ZIP作成完了: %d bytes", buf.Len())

	h.logger.CompleteOperation(tracker, map[string]interface{}{
		"success_count": len(result.Results),
		"error_count":   len(result.Errors),
		"zip_size":      buf.Len(),
	})

	// ZIPファイルを返す
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("converted_images_%s.zip", timestamp)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
	log.Printf("[ConvertBatchZip] ZIPダウンロード送信完了: %s", filename)
}
