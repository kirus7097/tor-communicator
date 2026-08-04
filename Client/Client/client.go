package Client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"golang.org/x/net/proxy"

	"torchat/Crypto"
)

// klient przechowuje informacje o polaczeniu
// bufor do wygodnego czytania danych z polaczeniu
// prywany klucz
// publiczny klucz
type Client struct {
	Conn       net.Conn
	Reader     *bufio.Reader
	PublicKey  *[32]byte
	PrivateKey *[32]byte
}

func Connect(server string) (*Client, error) {
	pub, priv, err := Crypto.LoadOrCreateKeypair()
	if err != nil {
		return nil, err
	}

	dialer, err := proxy.SOCKS5(
		"tcp",
		"127.0.0.1:9050",
		nil,
		proxy.Direct,
	)
	if err != nil {
		return nil, err
	}

	raw, err := dialer.Dial(
		"tcp",
		server,
	)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(
		raw,
		&tls.Config{
			ServerName: strings.Split(server, ":")[0],

			InsecureSkipVerify: true,
		},
	)

	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:       tlsConn,
		Reader:     bufio.NewReader(tlsConn),
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func (c *Client) Send(text string) (string, error) {
	_, err := c.Conn.Write(
		[]byte(text),
	)
	if err != nil {
		return "", err
	}

	return c.Reader.ReadString('\n')
}

func (c *Client) EncryptAndSend(
	target string,
	message string,
) error {
	pub, err := c.GetPublicKey(target)
	if err != nil {
		return err
	}

	encrypted, err := Crypto.EncryptMessage(
		[]byte(message),
		pub,
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
