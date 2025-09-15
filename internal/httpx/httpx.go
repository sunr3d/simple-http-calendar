package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func IsJSON(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if ct == "application/json" {
		return true
	}

	return strings.HasPrefix(ct, "application/json;")
}

func WriteJSON(w http.ResponseWriter, code int, v any) error {
	buff, err := json.Marshal(v)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrJSONMarshal, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(buff); err != nil {
		return fmt.Errorf("%w: %v", ErrWriteBody, err)
	}

	return nil
}

func HttpError(w http.ResponseWriter, code int, message string) error {
	return WriteJSON(w, code, map[string]string{"error": message})
}