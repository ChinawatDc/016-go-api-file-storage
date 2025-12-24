package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ChinawatDc/016-go-api-file-storage/internal/storage"
	"github.com/ChinawatDc/016-go-api-file-storage/internal/utils"
)

type FileHandler struct {
	Store         storage.Storage
	AllowedExt    map[string]bool
	MaxUploadMB   int64
	DefaultPrefix string
}

func NewFileHandler(store storage.Storage, allowed map[string]bool, maxMB int64, defaultPrefix string) *FileHandler {
	return &FileHandler{
		Store: store, AllowedExt: allowed, MaxUploadMB: maxMB,
		DefaultPrefix: defaultPrefix,
	}
}

func (h *FileHandler) maxBytes() int64 {
	return h.MaxUploadMB * 1024 * 1024
}

func (h *FileHandler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "missing file field 'file'"})
		return
	}

	if fh.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "empty file"})
		return
	}
	if fh.Size > h.maxBytes() {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "file too large"})
		return
	}

	ext := utils.ExtLower(fh.Filename)
	if !utils.IsAllowedExt(ext, h.AllowedExt) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "file extension not allowed", "ext": ext})
		return
	}

	safeName, err := utils.BuildSafeFilename(fh.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "filename error"})
		return
	}

	prefix := strings.TrimSpace(c.Query("prefix"))
	if prefix == "" {
		prefix = h.DefaultPrefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	key := prefix + safeName

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "open file failed"})
		return
	}
	defer f.Close()

	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	ctx := requestCtx(c)
	fi, err := h.Store.Put(ctx, storage.PutInput{
		Key:         key,
		Body:        f,
		Size:        fh.Size,
		ContentType: ct,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "upload failed", "error": err.Error()})
		return
	}

	expMin := int64(15)
	if v := c.Query("expire_min"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil && n > 0 {
			expMin = n
		}
	}
	u, _ := h.Store.GetURL(ctx, fi.Key, time.Duration(expMin)*time.Minute)
	fi.URL = u

	c.JSON(http.StatusOK, gin.H{"success": true, "file": fi})
}

func (h *FileHandler) GetURL(c *gin.Context) {
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "key required"})
		return
	}
	expMin := int64(15)
	if v := c.Query("expire_min"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil && n > 0 {
			expMin = n
		}
	}

	u, err := h.Store.GetURL(requestCtx(c), key, time.Duration(expMin)*time.Minute)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "get url failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "url": u})
}

func (h *FileHandler) List(c *gin.Context) {
	prefix := strings.TrimSpace(c.Query("prefix"))
	if prefix == "" {
		prefix = h.DefaultPrefix
	}
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	files, err := h.Store.List(requestCtx(c), prefix, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "list failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "files": files})
}

func (h *FileHandler) Delete(c *gin.Context) {
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "key required"})
		return
	}

	if err := h.Store.Delete(requestCtx(c), key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "delete failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func requestCtx(c *gin.Context) context.Context {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	_ = cancel
	return ctx
}
