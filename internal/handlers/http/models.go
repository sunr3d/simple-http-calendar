package httphandlers

type createEventReq struct {
	UserID   int64  `json:"user_id"`
	Date     string `json:"date"`
	Event    string `json:"event"`
	Reminder bool   `json:"reminder"`
}

type updateEventReq struct {
	EventID  string `json:"event_id"`
	UserID   int64  `json:"user_id"`
	Date     string `json:"date"`
	Event    string `json:"event"`
	Reminder bool   `json:"reminder"`
}

type deleteEventReq struct {
	EventID string `json:"event_id"`
}
