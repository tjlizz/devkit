package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devkit/server/internal/middleware"

	"github.com/google/uuid"
)

// maxArtifactUploadSize caps a single uploaded artifact (e.g. a source archive).
const maxArtifactUploadSize = 200 << 20 // 200 MB

type appArtifact struct {
	ID             int64  `json:"id"`
	AppID          int64  `json:"appId"`
	FileName       string `json:"fileName"`
	SizeBytes      int64  `json:"sizeBytes"`
	ContentType    string `json:"contentType"`
	ChecksumSHA256 string `json:"checksumSha256"`
	CreatedAt      string `json:"createdAt"`
}

type artifactMeta struct {
	ID             int64
	AppID          int64
	FileName       string
	StoredName     string
	SizeBytes      int64
	ContentType    string
	ChecksumSHA256 string
	CreatedAt      time.Time
}

// ListAppArtifacts returns the artifacts attached to one of the developer's apps.
func (h *Auth) ListAppArtifacts(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	if _, err := h.readDeveloperApp(r, appID, developerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "read developer app for artifacts", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}

	metas, err := h.listArtifactsRaw(r, appID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list app artifacts", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list artifacts")
		return
	}

	artifacts := make([]appArtifact, 0, len(metas))
	for _, meta := range metas {
		artifacts = append(artifacts, toAppArtifact(meta))
	}

	writeJSON(w, http.StatusOK, map[string]any{"artifacts": artifacts})
}

// UploadAppArtifact stores a downloadable artifact for one of the developer's apps.
func (h *Auth) UploadAppArtifact(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	if _, err := h.readDeveloperApp(r, appID, developerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "read developer app for upload", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxArtifactUploadSize+1024)
	file, header, err := r.FormFile("artifact")
	if err != nil {
		writeError(w, http.StatusBadRequest, "artifact file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxArtifactUploadSize+1))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read artifact upload", "error", err)
		writeError(w, http.StatusInternalServerError, "could not upload artifact")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "artifact file is required")
		return
	}
	if len(data) > maxArtifactUploadSize {
		writeError(w, http.StatusBadRequest, "artifact must be 200 MB or smaller")
		return
	}

	fileName := sanitizeArtifactFileName(header.Filename)
	if fileName == "" {
		writeError(w, http.StatusBadRequest, "artifact file name is invalid")
		return
	}

	if err := os.MkdirAll(h.config.ArtifactUploadDir, 0755); err != nil {
		h.logger.ErrorContext(r.Context(), "create artifact upload directory", "error", err)
		writeError(w, http.StatusInternalServerError, "could not upload artifact")
		return
	}

	storedName := fmt.Sprintf("app-%d-%s", appID, uuid.NewString())
	storedPath := filepath.Join(h.config.ArtifactUploadDir, storedName)
	if err := os.WriteFile(storedPath, data, 0644); err != nil {
		h.logger.ErrorContext(r.Context(), "write artifact upload", "name", header.Filename, "error", err)
		writeError(w, http.StatusInternalServerError, "could not upload artifact")
		return
	}

	contentType := http.DetectContentType(data)
	if contentType == "application/octet-stream" && header.Header.Get("Content-Type") != "" {
		contentType = header.Header.Get("Content-Type")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := h.db.ExecContext(r.Context(),
		`INSERT INTO app_artifacts(app_id, file_name, stored_name, size_bytes, content_type, checksum_sha256)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		appID, fileName, storedName, int64(len(data)), contentType, sha256Hex(data))
	if err != nil {
		_ = os.Remove(storedPath)
		h.logger.ErrorContext(r.Context(), "insert artifact", "error", err)
		writeError(w, http.StatusInternalServerError, "could not record artifact")
		return
	}
	artifactID, err := result.LastInsertId()
	if err != nil {
		_ = os.Remove(storedPath)
		h.logger.ErrorContext(r.Context(), "read artifact id", "error", err)
		writeError(w, http.StatusInternalServerError, "could not record artifact")
		return
	}

	meta, err := h.readArtifact(r, artifactID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read uploaded artifact", "error", err)
		writeError(w, http.StatusInternalServerError, "could not upload artifact")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]appArtifact{"artifact": toAppArtifact(meta)})
}

// DeleteAppArtifact removes an artifact owned by the developer's app.
func (h *Auth) DeleteAppArtifact(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	artifactID, ok := artifactIDFromPath(w, r)
	if !ok {
		return
	}

	meta, err := h.readArtifact(r, artifactID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "artifact not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read artifact for delete", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete artifact")
		return
	}
	if meta.AppID != appID {
		writeError(w, http.StatusNotFound, "artifact not found")
		return
	}
	// Verify the app belongs to the developer.
	if _, err := h.readDeveloperApp(r, appID, developerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "read developer app for delete artifact", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete artifact")
		return
	}

	if _, err := h.db.ExecContext(r.Context(),
		"DELETE FROM app_artifacts WHERE id = ? AND app_id = ?", artifactID, appID); err != nil {
		h.logger.ErrorContext(r.Context(), "delete artifact", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete artifact")
		return
	}
	_ = os.Remove(filepath.Join(h.config.ArtifactUploadDir, meta.StoredName))

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// DownloadAppArtifact serves an artifact to an entitled buyer. It accepts the
// short-lived signed delivery token (from GetDelivery) so the file can be
// fetched without a second JWT round-trip.
func (h *Auth) DownloadAppArtifact(w http.ResponseWriter, r *http.Request) {
	artifactID, ok := artifactIDFromPath(w, r)
	if !ok {
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeError(w, http.StatusUnauthorized, "delivery token required")
		return
	}

	entID, ok := verifyDeliveryToken(r, token, h.config.JWTSecret, h.now)
	if !ok {
		writeError(w, http.StatusUnauthorized, "delivery token is invalid or expired")
		return
	}

	ent, err := h.readEntitlement(r, entID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusForbidden, "entitlement not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read entitlement for download", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read entitlement")
		return
	}
	if ent.Status != "active" {
		writeError(w, http.StatusForbidden, "entitlement is not active")
		return
	}

	meta, err := h.readArtifact(r, artifactID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "artifact not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read artifact for download", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read artifact")
		return
	}
	if meta.AppID != ent.AppID {
		writeError(w, http.StatusForbidden, "artifact does not belong to this app")
		return
	}

	storedPath := filepath.Join(h.config.ArtifactUploadDir, meta.StoredName)
	file, err := os.Open(storedPath)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "open artifact file", "id", meta.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "could not open artifact")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(meta.SizeBytes, 10))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+sanitizeArtifactFileName(meta.FileName)+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, meta.FileName, meta.CreatedAt, file)
}

func (h *Auth) readArtifact(r *http.Request, artifactID int64) (artifactMeta, error) {
	var meta artifactMeta
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, app_id, file_name, stored_name, size_bytes, content_type, checksum_sha256, created_at
		 FROM app_artifacts WHERE id = ?`, artifactID).
		Scan(&meta.ID, &meta.AppID, &meta.FileName, &meta.StoredName,
			&meta.SizeBytes, &meta.ContentType, &meta.ChecksumSHA256, &meta.CreatedAt)
	return meta, err
}

func (h *Auth) listArtifactsRaw(r *http.Request, appID int64) ([]artifactMeta, error) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, app_id, file_name, stored_name, size_bytes, content_type, checksum_sha256, created_at
		 FROM app_artifacts WHERE app_id = ? ORDER BY id DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := make([]artifactMeta, 0)
	for rows.Next() {
		var meta artifactMeta
		if err := rows.Scan(&meta.ID, &meta.AppID, &meta.FileName, &meta.StoredName,
			&meta.SizeBytes, &meta.ContentType, &meta.ChecksumSHA256, &meta.CreatedAt); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, meta)
	}
	return artifacts, rows.Err()
}

func toAppArtifact(meta artifactMeta) appArtifact {
	return appArtifact{
		ID:             meta.ID,
		AppID:          meta.AppID,
		FileName:       meta.FileName,
		SizeBytes:      meta.SizeBytes,
		ContentType:    meta.ContentType,
		ChecksumSHA256: meta.ChecksumSHA256,
		CreatedAt:      meta.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// verifyDeliveryToken validates an HMAC-signed delivery token and returns the
// bound entitlement ID. Copies the format produced by signDeliveryToken.
func verifyDeliveryToken(r *http.Request, token string, secret string, now func() time.Time) (int64, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, false
	}
	// Layout: "<entID>:<expiryUnix>:<32 raw HMAC bytes>". The signature bytes
	// may themselves contain ':', so split on the first two separators only.
	parts := strings.SplitN(string(raw), ":", 3)
	if len(parts) != 3 {
		return 0, false
	}
	entID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || entID <= 0 {
		return 0, false
	}
	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	expiresAt := time.Unix(expiresAtUnix, 0)
	if now().After(expiresAt) {
		return 0, false
	}

	payload := fmt.Sprintf("%d:%d", entID, expiresAtUnix)
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return 0, false
	}
	signature := mac.Sum(nil)

	provided := []byte(parts[2])
	if len(provided) != len(signature) {
		return 0, false
	}
	if subtle.ConstantTimeCompare(provided, signature) != 1 {
		return 0, false
	}
	return entID, true
}

func sanitizeArtifactFileName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return name
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func artifactIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("artifactId"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid artifact id")
		return 0, false
	}
	return id, true
}
