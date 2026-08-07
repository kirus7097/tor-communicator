package CLI

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"torchat/Client"
)

func Run(c *Client.Client) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Tor Communicator")
	fmt.Println("----------------")
	fmt.Println("*REGISTER <username> <password>")
	fmt.Println("*LOGIN <username> <password>")
	fmt.Println("*MSG <target> <message>")
	fmt.Println("*LOGOUT")
	fmt.Println("*READ")
	fmt.Println()

	for {

		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(
				"Input closed:",
				err,
			)
			return
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)

		switch parts[0] {

		case "*REGISTER":

			if len(parts) != 3 {
				fmt.Println(
					"Usage: *REGISTER <username> <password>",
				)
				continue
			}

			reply, err := c.Register(
				parts[1],
				parts[2],
			)
			if err != nil {

				fmt.Println(
					"Error:",
					err,
				)

				continue
			}

			fmt.Print(reply)

		case "*LOGIN":
			if len(parts) != 3 {
				fmt.Println(
					"Usage: *LOGIN <username> <password>",
				)
				continue
			}

			reply, err := c.Login(
				parts[1],
				parts[2],
			)
			if err != nil {

				fmt.Println(
					"Error:",
					err,
				)

				continue
			}

			fmt.Print(reply)

		case "*MSG":

			if len(parts) < 3 {

				fmt.Println(
					"Usage: *MSG <target> <message>",
				)

				continue
			}

			target := parts[1]

			message := strings.Join(
				parts[2:],
				" ",
			)

			err := c.SendMessage(
				target,
				message,
			)
			if err != nil {

				fmt.Println(
					"Could not send:",
					err,
				)

				continue
			}

			fmt.Println(
				"Message sent",
			)

		case "*READ":

			messages, err := c.ReadMessages()
			if err != nil {

				fmt.Println(
					"Read failed:",
					err,
				)

				continue
			}

			if len(messages) == 0 {

				fmt.Println(
					"No messages",
				)

				continue
			}

			for _, msg := range messages {

				text, err := c.DecryptMessage(
					msg,
				)
				if err != nil {

					fmt.Println(
						"Decrypt failed:",
						err,
					)

					continue
				}

				fmt.Printf(
					"%s: %s\n",
					msg.Sender,
					text,
				)
			}

		case "*LOGOUT":

			reply, err := c.Logout()
			if err != nil {

				fmt.Println(err)

				continue
			}

			fmt.Print(reply)

		default:

			fmt.Println(
				"Unknown command",
			)
		}
	}
}
