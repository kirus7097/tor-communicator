package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	api "torchat/API"
	"torchat/CLI"
	"torchat/Client"
)

const server = "5jppyjmpqxcqmq4y27lw5ld24kfktlb6tn4rbzhspqyhdw7xeqrfrnid.onion:9090"

func main() {
	fmt.Println("Starting Torchat...")

	client, err := Client.Connect(server)
	if err != nil {
		fmt.Println(
			"Connection error:",
			err,
		)
		return
	}

	fmt.Println("Connected")

	http.HandleFunc("/api", api.ApiHandler)

	go func() {
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("Couldn't create a local server for API:", err)
		}
	}()

	startGui := exec.Command("python3", "gui.py")

	startGui.Stdout = os.Stdout
	startGui.Stderr = os.Stderr

	err = startGui.Run()
	if err != nil {
		fmt.Println("Couldn't start GUI. Entering CLI mode:", err)
	}

	CLI.Run(client)
}
