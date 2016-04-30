package main

import (
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"strings"
)

var config *ssh.ServerConfig

func startSshServer() {
	config = &ssh.ServerConfig{
		ServerVersion:     "SSH-2.0-BASKET-OF-KITTENS",
		PublicKeyCallback: publicKeyCallback,
	}

	pbytes, err := ioutil.ReadFile("/var/www/app/.ssh/id_rsa")
	if err != nil {
		// Handle the error
	}

	pkey, err := ssh.ParsePrivateKey(pbytes)
	if err != nil {
		// Handle the error
	}

	// Add the private key to the ServerConfig
	config.AddHostKey(pkey)

	listener, err := net.Listen("tcp", "[::]:5000")
	if err != nil {
		// Handle the error
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Handle the error
		}

		go func(c net.Conn) {
			_, channels, requests, err := ssh.NewServerConn(c, config)
			if err != nil {
				// Handle the error
			}

			// Discard all requests
			go ssh.DiscardRequests(requests)

			// Handle the connections
			handle(channels)

			// Close the connection
			conn.Close()
		}(conn)
	}

	wg.Done()
}

func publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	// Make sure the user has an active connection
	if _, ok := connections[conn.User()]; !ok {
		return nil, errors.New("User does not exist")
	}

	// Now make sure the key exists
	k, ok := keys[conn.User()]
	if !ok {
		return nil, errors.New("Key does not exist for user")
	}

	// Split the key up into parts to make sure it's the correct format.
	parts := strings.Split(k, " ")
	if len(parts) < 2 {
		return nil, errors.New("Invalid public key format")
	}

	// Encode public key to base64 for parsing
	encoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("Could not decode key")
	}

	// Parse public key
	pk, err := ssh.ParsePublicKey([]byte(encoded))
	if err != nil {
		return nil, err
	}

	// Make sure the key types match
	if key.Type() != pk.Type() {
		return nil, errors.New("Key types do not match")
	}

	kbytes := key.Marshal()
	pbytes := pk.Marshal()

	// Make sure the key lengths match
	if len(kbytes) != len(pbytes) {
		return nil, errors.New("Keys do not match")
	}

	// Make sure every byte of the key matches up
	for i, b := range kbytes {
		if b != pbytes[i] {
			return nil, errors.New("Keys do not match")
		}
	}

	// If we got this far, no issues were found!
	connections[conn.User()].WriteJSON(&Message{
		Type:    "ALERT",
		Id:      conn.User(),
		Message: "Your request has been authorized",
	})

	// If this were a real application we'd want to actually do some authentication
	// like setting it in a session and everything else. I'll just put this here
	// for you to do yourself: TODO

	return nil, nil
}

func handle(channels <-chan ssh.NewChannel) {
	for ch := range channels {
		// Reject all connections of a different type than "session"
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "Unknown channel type")
			continue
		}

		// Accept the channel
		channel, _, err := ch.Accept()
		if err != nil {
			// Handle the error
		}

		// Let the user know that they've been authorized in the shell.
		channel.Write([]byte("Authorized! "))
		channel.Close()

	}
}
