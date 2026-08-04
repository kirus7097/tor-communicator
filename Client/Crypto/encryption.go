package Crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

func GenerateKeypair() (pub, priv *[32]byte, err error) {
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	err = os.WriteFile(
		"private.key",
		priv[:],
		0o600,
	)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(
		"Public key:",
		hex.EncodeToString(pub[:]),
	)

	return pub, priv, nil
}

func LoadOrCreateKeypair() (
	pub *[32]byte,
	priv *[32]byte,
	err error,
) {
	data, err := os.ReadFile("private.key")

	if err == nil {

		if len(data) != 32 {
			return nil, nil,
				fmt.Errorf(
					"invalid private key length",
				)
		}

		priv = new([32]byte)

		copy(
			priv[:],
			data,
		)

		pub = new([32]byte)

		curve25519.ScalarBaseMult(
			pub,
			priv,
		)

		return pub, priv, nil
	}

	if !os.IsNotExist(err) {
		return nil, nil, err
	}

	return GenerateKeypair()
}

func EncryptMessage(
	plaintext []byte,
	recipientPub *[32]byte,
	senderPriv *[32]byte,
) (string, error) {
	var nonce [24]byte

	_, err := rand.Read(
		nonce[:],
	)
	if err != nil {
		return "", err
	}

	encrypted := box.Seal(
		nonce[:],
		plaintext,
		&nonce,
		recipientPub,
		senderPriv,
	)

	return hex.EncodeToString(
		encrypted,
	), nil
}

func DecryptMessage(
	ciphertext string,
	senderPub *[32]byte,
	receiverPriv *[32]byte,
) (string, error) {
	data, err := hex.DecodeString(
		ciphertext,
	)
	if err != nil {
		return "", err
	}

	if len(data) < 24 {
		return "", fmt.Errorf(
			"ciphertext too short",
		)
	}

	var nonce [24]byte

	copy(
		nonce[:],
		data[:24],
	)

	message, ok := box.Open(
		nil,
		data[24:],
		&nonce,
		senderPub,
		receiverPriv,
	)

	if !ok {
		return "", fmt.Errorf(
			"decryption failed",
		)
	}

	return string(message), nil
}
