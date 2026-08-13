package API

import (
	"encoding/json"
	"net/http"

	"torchat/Client"
)

type loginRequest struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type establishConnectionRequest struct {
	Target string `json:"target"`
}

type sendMessageRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
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

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleSendMessage(w, r)
	case http.MethodGet:
		handleGetMessages(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var req sendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := client.SendMessage(req.Target, req.Message); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"target":  req.Target,
		"message": req.Message,
	})
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		writeError(w, http.StatusBadRequest, "missing target query parameter")
		return
	}

	messages, err := client.GetMessages(target)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]interface{}{
		"target":   target,
		"messages": messages,
	})
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req establishConnectionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := client.GetPublicKey(req.Target); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"target": req.Target,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

		var req loginRequest

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
