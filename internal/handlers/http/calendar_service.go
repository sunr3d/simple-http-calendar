package http_handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/httpx"
)

func (h *httpHandler) createEvent(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("op", "handlers.createEvent"))

	logger.Info("получен запросна создание эвента")

	var req createEventReq // TODO: in models

	decoder := json.NewDecoder(r.Body)
	decoder.DissalowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		logger.Error("некорректный JSON", zap.Error(err))
		_ = httpx.HttpError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	// TODO: Продолжить логику

}

func (h *httpHandler) updateEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
}

func (h *httpHandler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
}

func (h *httpHandler) getDayEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
}

func (h *httpHandler) getWeekEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
}

func (h *httpHandler) getMonthEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
}
