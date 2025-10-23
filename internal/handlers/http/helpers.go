package httphandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sunr3d/simple-http-calendar/internal/handlers/validators"
	"github.com/sunr3d/simple-http-calendar/models"
)

func decodeBody(r *http.Request, dst any) error {
	ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}

		switch payload := dst.(type) {
		case *createEventReq:
			uid, _ := strconv.ParseInt(strings.TrimSpace(r.Form.Get("user_id")), 10, 64)
			payload.UserID = uid
			payload.Date = strings.TrimSpace(r.Form.Get("date"))
			payload.Event = r.Form.Get("event")
			payload.Reminder = r.Form.Get("reminder") == "true"
		case *updateEventReq:
			uid, _ := strconv.ParseInt(strings.TrimSpace(r.Form.Get("user_id")), 10, 64)
			payload.EventID = strings.TrimSpace(r.Form.Get("event_id"))
			payload.UserID = uid
			payload.Date = strings.TrimSpace(r.Form.Get("date"))
			payload.Event = r.Form.Get("event")
			payload.Reminder = r.Form.Get("reminder") == "true"
		case *deleteEventReq:
			payload.EventID = strings.TrimSpace(r.Form.Get("event_id"))
		default:
			return fmt.Errorf("неподдерживаемый payload")
		}

		return nil
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func parseQuery(r *http.Request) (models.EventsByDay, bool) {
	uidStr := strings.TrimSpace(r.URL.Query().Get("user_id"))
	dateStr := strings.TrimSpace(r.URL.Query().Get("date"))

	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		return models.EventsByDay{}, false
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return models.EventsByDay{}, false
	}

	filter := models.EventsByDay{UserID: uid, Day: date}
	if err := validators.ValidateFilter(filter); err != nil {
		return models.EventsByDay{}, false
	}

	return filter, true
}
