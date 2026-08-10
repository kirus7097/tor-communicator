package Client

import (
	"encoding/hex"
	"fmt"
	"strings"

	"torchat/Crypto"
)

type Message struct {
	Sender     string
	Ciphertext string
}

func (c *Client) GetPublicKey(
	username string,
) (*[32]byte, error) {
	reply, err := c.Send(
		fmt.Sprintf(
			"*GETKEY %s\n",
			username,
		),
	)
	if err != nil {
		return nil, err
	}

	keyBytes, err := hex.DecodeString(
		strings.TrimSpace(reply),
	)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != 32 {
		return nil,
			fmt.Errorf(
				"invalid public key",
			)
	}

	var key [32]byte

	copy(
		key[:],
		keyBytes,
	)

	return &key, nil
}

func (c *Client) Register(
	username string,
	password string,
) (string, error) {
	pubHex := fmt.Sprintf(
		"%x",
		c.PublicKey[:],
	)

	return c.Send(
		fmt.Sprintf(
			"*REGISTER %s %s %s\n",
			username,
			password,
			pubHex,
		),
	)
}

func (c *Client) Login(
	username string,
	password string,
) (string, error) {
	response, err := c.Send(
		fmt.Sprintf(
			"*LOGIN %s %s\n",
			username,
			password,
		),
	)

	if err != nil {
		return "", err
	}

	c.Username = username
	return response, nil
}

func (c *Client) Logout() (string, error) {
	response, err := c.Send(
		"*LOGOUT\n",
	)

	if err != nil {
		return "", err
	}

	c.Username = ""
	return response, nil
}

func (c *Client) SendMessage(
	target string,
	message string,
) error {
	receiverKey, err := c.GetPublicKey(
		target,
	)
	if err != nil {
		return err
	}

	encrypted, err := Crypto.EncryptMessage(
		[]byte(message),
		receiverKey,
		c.PrivateKey,
	)
	if err != nil {
		return err
	}

	_, err = c.Send(
		fmt.Sprintf(
			"*MSG %s %s\n",
			target,
			encrypted,
		),
	)

	return err
}

func (c *Client) ReadMessages() ([]Message, error) {
	_, err := c.Conn.Write(
		[]byte("*READ\n"),
	)
	if err != nil {
		return nil, err
	}

	var messages []Message

	for {

		line, err := c.Reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		if line == "END" {
			break
		}

		parts := strings.SplitN(
			line,
			": ",
			2,
		)

		if len(parts) != 2 {
			continue
		}

		messages = append(
			messages,
			Message{
				Sender:     parts[0],
				Ciphertext: parts[1],
			},
		)
	}

	return messages, nil
}

func (c *Client) DecryptMessage(
	msg Message,
) (string, error) {
	senderKey, err := c.GetPublicKey(
		msg.Sender,
	)
	if err != nil {
		return "", err
	}

	return Crypto.DecryptMessage(
		msg.Ciphertext,
		senderKey,
		c.PrivateKey,
	)
}
