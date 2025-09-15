package http_handlers

import (
	"net/http"

	"go.uber.org/zap"
)

type httpHandler struct {
	svc    services.CalendarService
	logger *zap.Logger
}

func New(svc services.CalendarService, logger *zap.Logger) *httpHandler {
	return &httpHandler{svc: svc, logger: logger}
}

// FIXME: здесь наверное нужно будет что-то починить в запросах, пока "черновик"
func (h *httpHandler) RegisterCalendarHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /create_event", h.createEvent)
	mux.HandleFunc("POST /update_event/{event_id}", h.updateEvent)
	mux.HandleFunc("POST /delete_event/{event_id}", h.deleteEvent)
	mux.HandleFunc("GET /events_for_day/{date}", h.getDayEvents)
	mux.HandleFunc("GET /events_for_week/{date}", h.getWeekEvents)
	mux.HandleFunc("GET /events_for_month/{date}", h.getMonthEvents)
}
