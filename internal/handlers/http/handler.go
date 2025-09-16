package httphandlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/interfaces/services"
)

type Handler struct {
	svc    services.CalendarService
	logger *zap.Logger
}

func New(svc services.CalendarService, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterCalendarHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /create_event", h.createEvent)
	mux.HandleFunc("POST /update_event", h.updateEvent)
	mux.HandleFunc("POST /delete_event", h.deleteEvent)
	mux.HandleFunc("GET /events_for_day", h.getDayEvents)
	mux.HandleFunc("GET /events_for_week", h.getWeekEvents)
	mux.HandleFunc("GET /events_for_month", h.getMonthEvents)
}
