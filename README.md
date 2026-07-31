# tor-communicator

Client-server encrypted messaging application written in Go.  
This project is a hands-on exploration of applied cryptography, network programming, and security engineering.

The goal is to build a small secure communication system while learning how concepts such as password hashing, public-key cryptography, authenticated encryption, proxy routing, and secure protocol design work in practice.

This is a learning project — I am developing my Go skills and security knowledge by implementing these concepts rather than only studying them theoretically.

## Why I built this

I am a self-taught software developer aspiring to work in cybersecurity. I wanted a project that would force me to apply security principles in practice instead of only reading about concepts such as encryption, authentication, hashing, and network security.

Building a messaging application combines many security topics together:
- identity management
- password storage
- public-key cryptography
- authenticated encryption
- network communication
- anonymity and metadata protection

## Current features

### Authentication and user management

- **User registration**
  - Users can create accounts with a username and password
  - Users provide a public encryption key during registration

- **Password hashing with bcrypt**
  - Passwords are never stored in plaintext
  - Password hashes are generated using bcrypt with the default cost factor

- **SQLite database storage**
  - User information is stored locally
  - Parameterized SQL queries are used to reduce SQL injection risks

### End-to-end encrypted messaging

- **Client-side message encryption**
  - Messages are encrypted before being sent to the server
  - The server stores only ciphertext, not plaintext messages

- **Public-key cryptography**
  - Each client generates a Curve25519 key pair
  - Public keys are shared with the server
  - Private keys remain locally on the client machine

- **Authenticated encryption using NaCl box**
  - Encryption provides confidentiality and authentication
  - Each message uses a unique random nonce

### Networking

- **Concurrent server connections**
  - Each client connection is handled in its own goroutine

- **Tor/SOCKS5 routing support**
  - The client can route connections through a local Tor SOCKS5 proxy
  - The goal is to reduce exposure of client IP metadata

- **Simple command-based protocol**
  - Registration
  - Login
  - Logout
  - Sending messages
  - Reading messages
  - Fetching public keys

## Security model

The current design provides:

### Protected against:
- Plaintext password storage
- Server-side plaintext message storage
- Simple SQL injection attacks
- Passive observation of message contents by the server

### Still visible to the server:
- Usernames
- Who sends messages to whom
- Message timing
- Connection information
- Account activity

The server is currently trusted to correctly deliver messages and provide public keys.

## Roadmap

This project is actively in progress.

Planned improvements:

- [ ] Replace newline-based protocol with proper message framing
- [ ] Add message deletion/read status handling
- [ ] Improve key management and storage
- [ ] Add key verification to prevent malicious public-key replacement
- [ ] Add unit and integration tests
- [ ] Improve CLI user experience
- [ ] Add better error handling
- [ ] Improve anonymity protections
- [ ] Study and compare with production messaging protocols
