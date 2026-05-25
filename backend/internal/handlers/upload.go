package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	miniogo "github.com/minio/minio-go/v7"
	"student-marketplace/internal/config"
	"student-marketplace/internal/middleware"
)

type UploadHandler struct {
	minio *miniogo.Client
	cfg   *config.Config
}

func NewUploadHandler(mc *miniogo.Client, cfg *config.Config) *UploadHandler {
	return &UploadHandler{minio: mc, cfg: cfg}
}

func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	return h.do(c, "images", []string{"image/jpeg", "image/png", "image/webp", "image/gif"}, 10<<20)
}
func (h *UploadHandler) UploadAvatar(c *fiber.Ctx) error {
	return h.do(c, "avatars", []string{"image/jpeg", "image/png", "image/webp"}, 5<<20)
}
func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	return h.do(c, "files", nil, 50<<20)
}

func (h *UploadHandler) do(c *fiber.Ctx, folder string, allowed []string, maxBytes int64) error {
	if h.minio == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "storage not configured")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file required")
	}
	if fh.Size > maxBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge,
			fmt.Sprintf("max %d MB", maxBytes>>20))
	}

	ct := fh.Header.Get("Content-Type")
	if len(allowed) > 0 {
		ok := false
		for _, a := range allowed {
			if strings.HasPrefix(ct, a) { ok = true; break }
		}
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "file type not allowed")
		}
	}

	f, _ := fh.Open()
	defer f.Close()
	buf, _ := io.ReadAll(io.LimitReader(f, maxBytes))

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext == "" { ext = extFromMime(ct) }
	obj := fmt.Sprintf("%s/%s/%d%s", folder, userID.String(), time.Now().UnixNano(), ext)

	_, err = h.minio.PutObject(context.Background(), h.cfg.MinIO.Bucket, obj,
		bytes.NewReader(buf), int64(len(buf)),
		miniogo.PutObjectOptions{ContentType: ct})
	if err != nil {
		return fiber.NewError(500, "upload failed: "+err.Error())
	}

	var publicBase string
	if h.cfg.MinIO.PublicURL != "" {
		publicBase = h.cfg.MinIO.PublicURL
	} else {
		scheme := "http"
		if h.cfg.MinIO.UseSSL {
			scheme = "https"
		}
		publicBase = fmt.Sprintf("%s://%s", scheme, h.cfg.MinIO.Endpoint)
	}
	url := fmt.Sprintf("%s/%s/%s", publicBase, h.cfg.MinIO.Bucket, obj)
	return c.JSON(fiber.Map{"url": url, "name": fh.Filename, "size": fh.Size, "type": ct})
}

func extFromMime(m string) string {
	t := map[string]string{"image/jpeg": ".jpg", "image/png": ".png",
		"image/webp": ".webp", "image/gif": ".gif", "application/pdf": ".pdf",
		"video/mp4": ".mp4", "audio/mpeg": ".mp3"}
	for k, v := range t {
		if strings.HasPrefix(m, k) { return v }
	}
	return ".bin"
}
