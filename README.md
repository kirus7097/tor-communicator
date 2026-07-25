# tor-communicator
Client-server messaging application with encryption in Go language, developed as a hands-on experience in applied cryptography and network security. In the future, it is supposed to be able to route the connection via Tor in order to secure metadata and investigate end-to-end encryption between clients.

It is a learning project – I am learning Go language, network programming and security principles while building it.

## Why I built this
I am an self-thought software developer aspiring for a future in the field of cyber security. I was seeking something which would compel me to apply the principles of security which I was reading about rather than only having theoretical knowledge of TLS handshakes, hash functions for passwords, etc.

## Current features

- **TLS-encrypted connections** — TLS-only connections are allowed (based on cert/key)
- **User registration** — user can register his/her credentials (username + password)
- **Passwords are hashed using bcrypt** — no plaintext passwords are stored
- **SQLite user storage** — with parameterized queries, to avoid SQL-injection
- **Support for concurrent connections** — each connection is processed separately in separate goroutines
- **Logging could be provided**
- **Messages prefixed with username**

## Roadmap

This project is actively in progress. Planned next steps:
- [ ] Message framing (moving off newline-delimited text to a proper length-prefixed protocol)
- [ ] Client-to-client messaging
- [ ] End-to-end encryption for message contents
- [ ] Tor / SOCKS proxy integration for connection-level metadata protection
- [ ] Unit tests
- [ ] Basic CLI client polish

## Getting started

### Prerequisites

- [Go](https://go.dev/dl/) 1.20 or later
- A TLS certificate and key (see below)

### Generate a certificate (for local testing)

```bash
openssl req -x509 -newkey rsa:2048 -keyout server.key -out server.crt -days 365 -nodes
```

### Run the server

```bash
go run server.go <port>
# example:
go run server.go 9090
```

### Registering a user

Once the server is running, connect to it with `client.go`, and register:
```REGISTER <username> <password>```

## What I've learned so far

- An example of the TLS handshake process in practice, beyond just theory
- The reason for bcrypt and other types of password hashing and how it differs from regular hashing
- Techniques to mitigate SQL injection with the use of prepared statements 
- The fundamentals of concurrent connections management with the use of goroutines
- Committing sensitive information such as your private keys to Git as a security hazard

## Disclaimer
It is a learning exercise and **has not** gone through any kind of security audit yet. It is not meant for production purposes or for securing real anonymity/security. If you want a tried-and-tested method of anonymous communication, you should go for [Signal](https://signal.org/) or the [Tor Browser](https://www.torproject.org/).
