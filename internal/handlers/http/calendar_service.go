package httphandlers

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/handlers/validators"
	"github.com/sunr3d/simple-http-calendar/internal/httpx"
	"github.com/sunr3d/simple-http-calendar/models"
)

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "handlers.createEvent"))

	logger.Info("получен запрос на создание эвента")

	var req createEventReq

	if err := decodeBody(r, &req); err != nil {
		logger.Warn("некорректное тело запроса", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.Date), time.UTC)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректная дата, ожидается YYYY-MM-DD")
		return
	}

	event := models.Event{UserID: req.UserID, Date: day, Text: req.Event}
	if err := validators.ValidateCreatePayload(event); err != nil {
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.svc.CreateEvent(r.Context(), event)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": id})
}

func (h *Handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "UpdateEvent"))

	logger.Info("получен запрос на обновление события")

	var req updateEventReq

	if err := decodeBody(r, &req); err != nil {
		logger.Warn("некорректное тело запроса", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.Date), time.UTC)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректная дата, ожидается YYYY-MM-DD")
		return
	}

	event := models.Event{ID: req.EventID, UserID: req.UserID, Date: day, Text: req.Event}
	if err := validators.ValidateUpdate(event); err != nil {
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateEvent(r.Context(), event); err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": "ok"})
}

func (h *Handler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "DeleteEvent"))

	logger.Info("получен запрос на удаление события")

	var req deleteEventReq

	if err := decodeBody(r, &req); err != nil {
		logger.Warn("некорректное тело запроса", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	if err := validators.ValidateDelete(req.EventID); err != nil {
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteEvent(r.Context(), req.EventID); err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": "ok"})
}

func (h *Handler) getDayEvents(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "GetDayEvents"))

	logger.Info("получен запрос на получение событий за день")

	filter, ok := parseQuery(r)
	if !ok {
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	events, err := h.svc.GetEventsForDay(r.Context(), filter.UserID, filter.Day)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": events})
}

func (h *Handler) getWeekEvents(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "GetWeekEvents"))

	logger.Info("получен запрос на получение событий за неделю")

	filter, ok := parseQuery(r)
	if !ok {
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	events, err := h.svc.GetEventsForWeek(r.Context(), filter.UserID, filter.Day)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": events})
}

func (h *Handler) getMonthEvents(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", "GetMonthEvents"))

	logger.Info("получен запрос на получение событий за месяц")

	filter, ok := parseQuery(r)
	if !ok {
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	events, err := h.svc.GetEventsForMonth(r.Context(), filter.UserID, filter.Day)
	if err != nil {
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": events})
}
