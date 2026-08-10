package API

import (
	"encoding/json"
	"net/http"

	"torchat/Client"
)

type Request struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Target   string `json:"target"`
	Message  string `json:"message"`
}

func replyToApi(code int, errorInfo string) map[string]any {
	return map[string]any{
		"status":  "error",
		"code":    code,
		"message": errorInfo,
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(
		replyToApi(code, msg),
	)
}

func writeSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"data":   data,
	})
}

var client *Client.Client

func SetClient(c *Client.Client) {
	client = c
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if client == nil {
			writeError(w, 500, "Client not initialized")
			return
		}

		writeSuccess(w, map[string]any{
			"loggedIn": client.Username != "",
			"username": client.Username,
		})
		return

	case http.MethodPost:
		if client == nil {
			writeError(w, 500, "Client not initialized")
			return
		}

		var req Request

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, 400, "Invalid JSON")
			return
		}

		var response any

		switch req.Type {
		case "login":
			response, err = client.Login(
				req.Username,
				req.Password,
			)

		case "register":
			response, err = client.Register(
				req.Username,
				req.Password,
			)

		case "logout":
			response, err = client.Logout()

		case "send":
			err = client.SendMessage(
				req.Target,
				req.Message,
			)

			if err == nil {
				response = "sent"
			}

		case "messages":
			response, err = client.ReadMessages()

		default:
			writeError(
				w,
				400,
				"Unknown request type",
			)
			return
		}

		if err != nil {
			writeError(
				w,
				500,
				err.Error(),
			)
			return
		}

		writeSuccess(
			w,
			response,
		)
		return
	}

	writeError(
		w,
		http.StatusMethodNotAllowed,
		"Method not allowed",
	)
}
