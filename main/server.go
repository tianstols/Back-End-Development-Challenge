package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

//NewServer function that waits for connection
func NewServer() {
	counter := 1
	fmt.Println("You are server")
	clientManager := ClientManager{}
	clientManager.game = CreateGame(5, 5)
	responseChannel := make(chan Response)
	go HandleClientRequests(responseChannel, &clientManager)

	listen, err := net.Listen(Type, Host+":"+Port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Listening on " + Host + ":" + Port)
	defer listen.Close()

	for {
		connection, err := listen.Accept()
		if err != nil {
			fmt.Println("Could not accept connection: ", err.Error())
			continue
		}

		if len(clientManager.clients) == 4 {
			fmt.Println("Server full") //Wanted clients to connect to different lobbies when full :(
			connection.Close()
			continue
		}

		buf := make([]byte, 1024)
		len, _ := connection.Read(buf)
		name := string(buf[:len])
		fmt.Println(name + " connected.")

		client := CreateClient(counter, name, connection, responseChannel)
		clientManager.AddClient(client)
		client.SendMessage("ID")
		counter++

		go Lobby(client, &clientManager)
	}

}

//CreateClient creates a new client object
func CreateClient(counter int, name string, connection net.Conn, responseChannel chan Response) *Client {
	client := Client{counter, name, connection, responseChannel, Player{name, NotReady, Stats{}}}
	return &client
}

//HandleClientRequests processes the client requests
func HandleClientRequests(requests chan Response, cm *ClientManager) {
	for {
		request := <-requests
		client := cm.clients[request.ClientID]
		if client.id != request.ClientID {
			fmt.Println("Something is very wrong")
		}
		//fmt.Println(request.Text + " Time: " + request.Sent.String())
		if request.Text[0] == '/' {
			switch request.Text {
			case "/exit":
				client.clientCon.Close()
				delete(cm.clients, client.id)
			case "/ready":
				client.Ready()
			case "/notready":
				client.NotReady()
			case "/start":
				client.Start()
			default:
				client.SendMessage("Unknown command")
				continue
			}

		}
		cm.UpdateGameState(request)
	}
}

//SendMessage sends a message to a specific client
func (c *Client) SendMessage(message string) {
	bytes := SerialiseResponse(c.id, message)
	c.clientCon.Write(bytes)
}

//BroadcastMessage sends a message to all connected clients
func (cm *ClientManager) BroadcastMessage(message string) {
	for _, v := range cm.clients {
		v.SendMessage(message)
	}
}

//Reader reads messages from connection
func (c *Client) Reader(cm *ClientManager) {
	for {
		buf := make([]byte, 1024)
		len, err := c.clientCon.Read(buf)
		if err != nil {
			fmt.Println(c.name + " disconnected.")
			c.clientCon.Close()
			delete(cm.clients, c.id)
			return
		}
		c.ch <- DeserialiseResponse(buf[:len])
	}
}

//SerialiseResponse converts the Response object to a bytes array
func SerialiseResponse(clientID int, message string) []byte {
	sent := time.Now()
	response := Response{clientID, message, sent}
	bytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Could not serialize message: " + err.Error())
	}
	return bytes
}

//DeserialiseResponse converts the json bytes to a Response object
func DeserialiseResponse(b []byte) Response {
	var response Response
	json.Unmarshal(b, &response)
	return response
}

//Ready marks the player as ready
func (c *Client) Ready() {
	c.player.State = Ready
}

//NotReady marks the player as not ready
func (c *Client) NotReady() {
	c.player.State = NotReady
}

//Start will
func (c *Client) Start() {
	c.player.State = Start
}

//DisconnectAllClients disconnects the clients after the game is finished
func (cm *ClientManager) DisconnectAllClients() {
	for i, v := range cm.clients {
		v.clientCon.Close()
		delete(cm.clients, i)
	}
}
