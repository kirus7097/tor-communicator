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

var client *Client.Client

func SetClient(c *Client.Client) {
	client = c
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
		return
	}

	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", 400)
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

		response = "sent"

	case "messages":
		response, err = client.ReadMessages()

	default:
		http.Error(w, "Unknown request", 400)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(response)
}
