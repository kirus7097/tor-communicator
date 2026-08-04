package main

import (
	"fmt"
	"os/exec"

	"torchat/CLI"
	"torchat/Client"
)

const server = "5jppyjmpqxcqmq4y27lw5ld24kfktlb6tn4rbzhspqyhdw7xeqrfrnid.onion:9090"

func main() {
	fmt.Println("Starting Torchat...")

	client, err := Client.Connect(
		server,
	)
	if err != nil {

		fmt.Println(
			"Connection error:",
			err,
		)

		return
	}

	fmt.Println("Connected")
	exec.Command("python", "gui.py").Run()
	CLI.Run(client)
}
