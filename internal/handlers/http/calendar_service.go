package httphandlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

	day, err := time.ParseInLocation("2006-01-02T15:04:05", strings.TrimSpace(req.Date), time.Local)
	if err != nil {
		logger.Warn("некорректная дата", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректная дата, ожидается YYYY-MM-DDTHH:MM:SS")
		return
	}

	event := models.Event{UserID: req.UserID, Date: day, Text: req.Event, Reminder: req.Reminder}
	if err := validators.ValidateCreatePayload(event); err != nil {
		logger.Warn("некорректные данные события", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.svc.CreateEvent(r.Context(), event)
	if err != nil {
		logger.Warn("ошибка при создании события", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	logger.Info("событие успешно создано", zap.String("event_id", id))
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

	day, err := time.ParseInLocation("2006-01-02T15:04:05", strings.TrimSpace(req.Date), time.Local)
	if err != nil {
		logger.Warn("некорректная дата", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректная дата, ожидается YYYY-MM-DDTHH:MM:SS")
		return
	}

	event := models.Event{ID: req.EventID, UserID: req.UserID, Date: day, Text: req.Event, Reminder: req.Reminder}
	if err := validators.ValidateUpdate(event); err != nil {
		logger.Warn("некорректные данные события", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateEvent(r.Context(), event); err != nil {
		logger.Warn("ошибка при обновлении события", zap.String("event_id", req.EventID), zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	logger.Info("событие успешно обновлено", zap.String("event_id", req.EventID))
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
		logger.Warn("некорректные данные события", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteEvent(r.Context(), req.EventID); err != nil {
		logger.Warn("ошибка при удалении события", zap.String("event_id", req.EventID), zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	logger.Info("событие успешно удалено", zap.String("event_id", req.EventID))
	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": "ok"})
}

func (h *Handler) getDayEvents(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, "GetDayEvents", h.svc.GetEventsForDay)
}

func (h *Handler) getWeekEvents(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, "GetWeekEvents", h.svc.GetEventsForWeek)
}

func (h *Handler) getMonthEvents(w http.ResponseWriter, r *http.Request) {
	h.getEvents(w, r, "GetMonthEvents", h.svc.GetEventsForMonth)
}

func (h *Handler) getEvents(
	w http.ResponseWriter,
	r *http.Request,
	op string,
	eventsFunc func(context.Context, int64, time.Time) ([]models.Event, error),
) {
	logger := h.logger.With(zap.String("component", "handler"), zap.String("op", op))

	filter, ok := parseQuery(r)
	if !ok {
		logger.Warn("некорректный запрос")
		_ = httpx.HTTPError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	logger.Info(
		fmt.Sprintf("получен запрос на получение событий %s для пользователя %d",
			op,
			filter.UserID),
	)

	events, err := eventsFunc(r.Context(), filter.UserID, filter.Day)
	if err != nil {
		logger.Warn("ошибка при получении событий", zap.Error(err))
		_ = httpx.HTTPError(w, http.StatusServiceUnavailable, "Сервис недоступен")
		return
	}

	logger.Info("события успешно получены", zap.String("user_id", strconv.FormatInt(filter.UserID, 10)))
	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"result": events})
}
