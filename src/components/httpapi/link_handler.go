// src/components/httpapi/link_handler.go
// Handles HTTP requests for health checks, link creation, redirects, and stats retrieval.
// Connects to: src/services/link_service.go, src/components/httpapi/response.go
// Created: 2026-06-28

package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/breakingthebot/url-shortener-api/src/services"
)

// LinkHandler wires HTTP endpoints to link service operations.
type LinkHandler struct {
	service services.LinkService
	logger  *slog.Logger
}

// NewLinkHandler constructs an HTTP handler for link operations.
func NewLinkHandler(service services.LinkService, logger *slog.Logger) LinkHandler {
	return LinkHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes binds all API and redirect endpoints to the provided mux.
func (h LinkHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthCheck)
	mux.HandleFunc("POST /api/v1/links", h.handleCreateLink)
	mux.HandleFunc("GET /api/v1/links/{code}", h.handleGetLinkStats)
	mux.HandleFunc("GET /{code}", h.handleRedirect)
}

// handleHealthCheck confirms that the HTTP service is reachable.
func (h LinkHandler) handleHealthCheck(writer http.ResponseWriter, _ *http.Request) {
	WriteJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

// handleCreateLink creates a new shortcode for a submitted URL.
func (h LinkHandler) handleCreateLink(writer http.ResponseWriter, request *http.Request) {
	var createLinkRequest CreateLinkRequest
	if err := json.NewDecoder(request.Body).Decode(&createLinkRequest); err != nil {
		WriteError(writer, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	link, created, err := h.service.CreateShortLink(
		request.Context(),
		createLinkRequest.OriginalURL,
		createLinkRequest.CustomCode,
	)
	if err != nil {
		h.writeServiceError(writer, err)
		return
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusCreated
	}

	WriteJSON(writer, statusCode, link)
}

// handleGetLinkStats returns the saved details and click count for a shortcode.
func (h LinkHandler) handleGetLinkStats(writer http.ResponseWriter, request *http.Request) {
	code := request.PathValue("code")
	link, err := h.service.GetLinkStats(request.Context(), code)
	if err != nil {
		h.writeServiceError(writer, err)
		return
	}

	WriteJSON(writer, http.StatusOK, link)
}

// handleRedirect resolves a shortcode to its original URL and issues the redirect.
func (h LinkHandler) handleRedirect(writer http.ResponseWriter, request *http.Request) {
	code := request.PathValue("code")
	originalURL, err := h.service.ResolveShortLink(request.Context(), code)
	if err != nil {
		h.writeServiceError(writer, err)
		return
	}

	http.Redirect(writer, request, originalURL, http.StatusTemporaryRedirect)
}

// writeServiceError maps domain failures into stable HTTP responses.
func (h LinkHandler) writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidURL):
		WriteError(writer, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrInvalidCustomCode):
		WriteError(writer, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrCustomCodeUnavailable):
		WriteError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrURLAlreadyShortened):
		WriteError(writer, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrLinkNotFound):
		WriteError(writer, http.StatusNotFound, "link not found")
	default:
		h.logger.Error("request failed", "error", err)
		WriteError(writer, http.StatusInternalServerError, "internal server error")
	}
}
