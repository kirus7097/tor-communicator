# tor-communicator

A TLS-encrypted client-server communication app written in Go, built as a hands-on learning project in applied cryptography and network security. The long-term goal is to route traffic through Tor for metadata protection, and to explore end-to-end encryption between clients.

This is an educational project — I'm learning Go, network programming, and security concepts as I build it, and documenting that process along the way.

## Why I built this

I'm a self-taught developer building toward a career in cybersecurity. I wanted a project that would force me to actually implement the security concepts I was reading about, rather than just read theory — things like TLS handshakes, password hashing, and safe database queries.

## Current features

- **TLS-encrypted connections** — the server only accepts connections secured with TLS (cert/key based)
- **User registration** — clients can register a username and password
- **Password hashing with bcrypt** — passwords are never stored in plaintext
- **SQLite user storage** — using parameterized queries to prevent SQL injection
- **Concurrent connection handling** — each client connection is handled in its own goroutine

## Roadmap

This project is actively in progress. Planned next steps:

- [ ] `LOGIN` command and session/auth token handling
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

- How TLS handshakes and certificate loading work in practice, not just in theory
- Why password hashing algorithms like bcrypt exist and how they differ from plain hashing
- How to prevent SQL injection using parameterized queries
- The basics of concurrent connection handling with goroutines
- Why committing secrets (like private keys) to version control is a real risk, not just a rule

## Disclaimer
This is a learning project and has **not** been security-audited. It is not intended for production use or to protect real anonymity/security needs. If you're looking for a battle-tested anonymous communication tool, use something like [Signal](https://signal.org/) or the [Tor Browser](https://www.torproject.org/) directly.
