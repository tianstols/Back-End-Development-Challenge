package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

//NewClient function that connects to server
func NewClient() {
	fmt.Println("You are client")
	fmt.Println("Enter your nickname:")
	channel := make(chan Response)

	reader := bufio.NewReader(os.Stdin)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	name = strings.Replace(name, "\r\n", "", -1)

	connection, err := net.Dial(Type, Host+":"+Port)
	if err != nil {
		println(err)
	}

	server := Server{0, connection, channel}
	go HandleServerResponse(channel, &server)
	connection.Write([]byte(name))
	server.Listen()
}

//Listen to server and writes to server
func (s *Server) Listen() {
	go s.Reader()
	s.Writer()
}

//Writer writes messages to connection
func (s *Server) Writer() {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			///fmt.Println("Server disconnected")
			s.serverCon.Close()
			return
		}
		message = strings.Replace(message, "\r\n", "", -1)
		if message != "" {
			bytes := SerialiseResponse(s.clientID, message)
			s.serverCon.Write([]byte(bytes))
		}

	}

}

//Reader reads messages from connection
func (s *Server) Reader() {
	buf := make([]byte, 1024)
	for {
		len, err := s.serverCon.Read(buf)
		if err != nil {
			fmt.Println("Server disconnected")
			s.serverCon.Close()
			os.Exit(1)
		}

		s.ch <- DeserialiseResponse(buf[:len])

	}
}

//HandleServerResponse processes the message from the server
func HandleServerResponse(serverResponse chan Response, s *Server) {
	for {
		response := <-serverResponse
		switch response.Text {
		case "ID":
			s.clientID = response.ClientID
			continue
		}
		fmt.Println(response.Text)
	}

}
