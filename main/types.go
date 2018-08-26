package main

import (
	"net"
	"time"
)

//Provides the connection info
const (
	Host = "localhost"
	Port = "3333"
	Type = "tcp"
)

//Player states
const (
	Ready     = 0
	NotReady  = 1
	Start     = 2
	Answering = 3
	Answered  = 4
)

//Game state
const (
	InLobby         = 0
	AwaitingAnswers = 1
	Busy            = 2
	Finished        = 3
)

//Player information for each client
type Player struct {
	Name  string
	State int
	Stats Stats
}

//ClientManager contains all the connections for the clients
type ClientManager struct {
	clients map[int]*Client
	game    *Game
}

//AddClient adds a new client to the client manager
func (cm *ClientManager) AddClient(c *Client) {
	if cm.clients == nil {
		cm.clients = make(map[int]*Client)
	}
	cm.clients[c.id] = c
}

//Client information
type Client struct {
	id        int
	name      string
	clientCon net.Conn
	ch        chan Response
	player    Player
}

//Server contains all clients and connection info
type Server struct {
	clientID  int
	serverCon net.Conn
	ch        chan Response
}

//Game stores the state of the game
type Game struct {
	CurrentRound int
	State        int
	Scoreboard   map[int]int
	answerCh     chan Response
	maxRounds    int
	maxTime      int
}

//Stats store all the player's stats
type Stats struct {
	Points          []int
	ResponseTimes   []time.Duration
	AvgResponseTime []time.Duration
	StdResponseTime []time.Duration
}

//Response of the player
type Response struct {
	ClientID int
	Text     string
	Sent     time.Time
}
