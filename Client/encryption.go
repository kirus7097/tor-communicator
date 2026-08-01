package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

// creating struct of type Message
// it stores who sent the message and the encrypted content
// TODO: revise myself more about structs in Go
type Message struct {
	Sender     string
	Ciphertext string
}

// generate both public and private key
// public key is shown to user because other people need it to encrypt messages for us
// private key is saved locally because only we should have access to it
//
// NOTE: permissions are 0600:
// owner can read and edit, nobody else can access it
// maybe change edit permission? user normally shouldn't manually edit private key
func GenerateKeypair() (pub, priv *[32]byte, err error) {
	pub, priv, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// saving private key as raw 32 bytes
	// this file is needed so we don't create a new identity every time program starts
	err = os.WriteFile("private.key", priv[:], 0o600)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Public key:", hex.EncodeToString(pub[:]))

	return pub, priv, nil
}

// check if private key already exists
// if it exists -> load it
// if not -> generate a new keypair
//
// KEYS ARE ALWAYS 32 BYTES
func LoadOrCreateKeypair() (pub, priv *[32]byte, err error) {
	// try to read existing private key
	data, err := os.ReadFile("private.key")

	if err == nil {

		// make sure file wasn't damaged or manually changed
		if len(data) != 32 {
			return nil, nil, fmt.Errorf(
				"private.key is corrupt: expected 32 bytes, got %d",
				len(data),
			)
		}

		// recreate private key from file
		priv = new([32]byte)
		copy(priv[:], data)

		// public key can be calculated from private key
		// no need to store public key separately
		pub = new([32]byte)
		curve25519.ScalarBaseMult(pub, priv)

		return pub, priv, nil
	}

	// if error is something different than "file doesn't exist"
	// then something else went wrong
	if !os.IsNotExist(err) {
		return nil, nil, err
	}

	// no private.key found -> create new identity
	return GenerateKeypair()
}

// nonce is responsible for making every encrypted message different
//
// example:
// "hello" encrypted once != "hello" encrypted second time
//
// without nonce, same messages would create the same ciphertext
// which leaks information
func EncryptMessage(
	plaintext []byte,
	recipientPub *[32]byte,
	senderPriv *[32]byte,
) (string, error) {
	var nonce [24]byte

	// generate random nonce for this specific message
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}

	// include nonce at beginning of encrypted message
	// because receiver needs it to decrypt
	encrypted := box.Seal(
		nonce[:],
		plaintext,
		&nonce,
		recipientPub,
		senderPriv,
	)

	// convert bytes to text because easier to send over network
	return hex.EncodeToString(encrypted), nil
}

// decrypt message using:
// - sender public key (to verify who sent it)
// - our private key (to decrypt it)
//
// NOTE:
// spelling "reciever" should probably be fixed to "receiver"
func decryptMessage(
	ciphertext string,
	senderPub *[32]byte,
	receiverPriv *[32]byte,
) (string, error) {
	// convert text back into bytes
	ciphertextHexDecoded, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// first 24 bytes are nonce
	if len(ciphertextHexDecoded) < 24 {
		return "", fmt.Errorf("ciphertext too short")
	}

	var nonce [24]byte
	copy(nonce[:], ciphertextHexDecoded[:24])

	// remove nonce and decrypt the actual message
	finalMessage, ok := box.Open(
		nil,
		ciphertextHexDecoded[24:],
		&nonce,
		senderPub,
		receiverPriv,
	)

	if !ok {
		return "", fmt.Errorf("could not decrypt message")
	}

	return string(finalMessage), nil
}
