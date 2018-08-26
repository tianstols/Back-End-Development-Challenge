package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

//Lobby adds a player to the first open lobby
func Lobby(c *Client, cm *ClientManager) {

	time.Sleep(time.Millisecond * 200)
	c.SendMessage("Welcome to the lobby.")
	c.SendMessage("Available commands: /ready, /unready, /start, /exit")
	c.Reader(cm)
}

//StartGame begins the game
func StartGame(cm *ClientManager) {
	rand.Seed(time.Now().UTC().UnixNano())
	totalPlayers := len(cm.clients)
	for i := 1; i <= cm.game.maxRounds; i++ {
		if len(cm.clients) < totalPlayers {
			cm.BroadcastMessage("Player disconnected. Ending game")
			return
		}
		cm.BroadcastMessage("Round " + strconv.Itoa(i) + " starting in 3")
		time.Sleep(time.Second)
		cm.BroadcastMessage("2")
		time.Sleep(time.Second)
		cm.BroadcastMessage("1")
		time.Sleep(time.Second)
		cm.BroadcastMessage("Start. Time limit of " + strconv.Itoa(cm.game.maxTime) + " seconds")

		correctAnswer := cm.GenerateEquation()

		startTime := time.Now()
		timer := time.NewTimer(time.Duration(cm.game.maxTime) * time.Second)
		cm.game.State = AwaitingAnswers
		cm.UpdatePlayerStates(Answering)
		stopRound := false
		playerAnswers := []Response{}
		totalAnswers := 0
		for {
			select {
			case <-timer.C:
				stopRound = true

				cm.BroadcastMessage("Timer expired")
				for _, v := range cm.clients {
					if v.player.State != Answered {
						expiredResponse := Response{v.id, "expired", time.Now()}
						playerAnswers = append(playerAnswers, expiredResponse)
						fmt.Println("Client not answered")
					}
				}
				cm.UpdatePlayerStates(Answered)

			case playerAnswer := <-cm.game.answerCh:
				cm.clients[playerAnswer.ClientID].player.State = Answered
				playerAnswers = append(playerAnswers, playerAnswer)
				totalAnswers++
				if totalAnswers == totalPlayers {
					stopRound = true
					timer.Stop()
				}
			}
			if stopRound {
				break
			}
		}
		totalAnswers = 0
		cm.game.State = Busy
		cm.BroadcastMessage("Round finished")
		cm.CalculatePoints(playerAnswers, correctAnswer, startTime)
		cm.UpdateScoreboard()
		cm.GenerateStats()
		//broadcast scoreboard and each player's stats
		cm.BroadcastScoreboard(i)
	}
	//Broadcast winner
	cm.BroadcastWinner()
	cm.ClearAllStates()
	cm.DisconnectAllClients()
	return
}

//GenerateEquation create the problem that needs to be solved and sends it to the players
func (cm *ClientManager) GenerateEquation() string {
	num1 := rand.Intn(10)
	num2 := rand.Intn(10)
	equation := strconv.Itoa(num1) + " x " + strconv.Itoa(num2)
	cm.BroadcastMessage(equation)
	return strconv.Itoa(num1 * num2)

}

//CalculatePoints calculates the points of each player
func (cm *ClientManager) CalculatePoints(pa []Response, correctAnswer string, startTime time.Time) {
	maxPoints := len(cm.clients)
	pa = OrderByReceived(pa)
	for _, v := range pa {
		if v.Text == correctAnswer {
			cm.clients[v.ClientID].player.Stats.Points = append(cm.clients[v.ClientID].player.Stats.Points, maxPoints)
			maxPoints--
		} else {
			cm.clients[v.ClientID].player.Stats.Points = append(cm.clients[v.ClientID].player.Stats.Points, 0)
		}

		responseTime := v.Sent.Sub(startTime)
		cm.clients[v.ClientID].player.Stats.ResponseTimes = append(cm.clients[v.ClientID].player.Stats.ResponseTimes, responseTime)
	}
}

//OrderByReceived will order the answers by first
func OrderByReceived(array []Response) []Response {
	for i := 0; i < len(array)-1; i++ {
		for j := 1; j < len(array); j++ {
			if array[j].Sent.Before(array[i].Sent) {
				temp := array[i]
				array[i] = array[j]
				array[j] = temp
			}
		}
	}
	return array
}

//UpdatePlayerStates updates the state of all the players
func (cm *ClientManager) UpdatePlayerStates(state int) {
	for _, v := range cm.clients {
		v.player.State = state
	}
}

//UpdateScoreboard updates the scoreboard for each round
func (cm *ClientManager) UpdateScoreboard() {
	for _, v := range cm.clients {
		playerScore := 0
		for i := range v.player.Stats.Points {
			playerScore += v.player.Stats.Points[i]
		}
		cm.game.Scoreboard[v.id] = playerScore
	}
}

//GenerateStats will generate the statistics for each player per round
func (cm *ClientManager) GenerateStats() {
	for _, v := range cm.clients {
		var sumResponse time.Duration
		count := 0
		for i := range v.player.Stats.ResponseTimes {
			sumResponse += v.player.Stats.ResponseTimes[i]
			count++
		}
		if count != 0 {
			v.player.Stats.AvgResponseTime = append(v.player.Stats.AvgResponseTime, sumResponse/(time.Duration(count)*time.Nanosecond))

			sumResponse = 0
			count = 0
			//Calculate standard deviation
			for _, j := range v.player.Stats.ResponseTimes {
				diff := j - v.player.Stats.AvgResponseTime[count]
				fl := float64(diff / time.Nanosecond)
				squared := math.Pow(fl, 2)
				ns := time.Duration(squared) * time.Nanosecond
				sumResponse += ns
				count++
			}
			variance := sumResponse / (time.Duration(count) * time.Nanosecond)

			stdDev := math.Sqrt(float64(variance / time.Nanosecond))

			v.player.Stats.StdResponseTime = append(v.player.Stats.StdResponseTime, time.Duration(stdDev)*time.Nanosecond)
		}

	}
}

//BroadcastScoreboard send the scoreboard to all the players
func (cm *ClientManager) BroadcastScoreboard(round int) {
	for _, v := range cm.clients {
		message := fmt.Sprintf("%s - Score: %d, response time: %v, avg. response time: %v, std. deviation: %v", v.name, cm.game.Scoreboard[v.id], v.player.Stats.ResponseTimes[round-1], v.player.Stats.AvgResponseTime[round-1],
			v.player.Stats.StdResponseTime[round-1])
		fmt.Println(message)
		cm.BroadcastMessage(message)
	}
}

//CreateGame creates a new game object
func CreateGame(rounds int, roundTime int) *Game {
	game := Game{0, InLobby, make(map[int]int), make(chan Response), rounds, roundTime}
	return &game
}

//BroadcastWinner determines the winner of the game and broadcasts it to all the players
func (cm *ClientManager) BroadcastWinner() {
	winnerID := -1
	score := -1
	for i, v := range cm.game.Scoreboard {
		if v > score {
			score = v
			winnerID = i
		}
	}
	winnerName := cm.clients[winnerID].name
	cm.BroadcastMessage(winnerName + " is the winner!")
}

//UpdateGameState will update the state of the game
func (cm *ClientManager) UpdateGameState(response Response) {
	switch cm.game.State {
	case InLobby:
		switch cm.clients[response.ClientID].player.State {
		case Ready:
			cm.clients[response.ClientID].SendMessage("You are ready")
			for _, v := range cm.clients {
				if v.player.State == NotReady {
					return
				}
			}
			cm.BroadcastMessage("All player ready, start the game")
			return
		case NotReady:
			cm.clients[response.ClientID].SendMessage("You are not ready")
			return
		case Start:
			start := false
			for _, v := range cm.clients {
				if v.player.State == NotReady {
					cm.clients[response.ClientID].SendMessage("Not all players are ready")
					return
				}
				if v.player.State == Start {
					start = true
				}
			}

			if len(cm.clients) <= 1 && start == true {
				cm.clients[response.ClientID].SendMessage("Not enough players")
				cm.clients[response.ClientID].player.State = Ready
				return
			}
		}

		cm.game.State = Busy
		go StartGame(cm)
	case AwaitingAnswers:
		switch cm.clients[response.ClientID].player.State {
		case Answering:
			cm.game.answerCh <- response
		case Answered:
			cm.clients[response.ClientID].SendMessage("Already answered")

		}

	case Busy:
		return
	case Finished:
	}
}

//ClearAllStates reverts the states and clears the data
func (cm *ClientManager) ClearAllStates() {
	cm.game = CreateGame(cm.game.maxRounds, cm.game.maxTime)
}
