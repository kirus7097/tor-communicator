package API

import (
	"encoding/json"
	"io"
	"net/http"
)

type loginOrRegisterInfo struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var data loginOrRegisterInfo

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
}
