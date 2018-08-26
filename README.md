# Back-End-Development-Challenge

The game was written in Golang. All the necessary files are inside the main folder.

## Design

### Communication

TCP was used for the communication between the server and the clients. The server will keep listening for connections, but will limit the connections to four at a time. Golang threads and channels were used to allow the server and clients to communicate anytime, but the server decides if the data from clients are to be used depending on the states of the other clients and the game state. The following diagram shows how the communication was implemented:

![Communication](https://github.com/tianstols/Back-End-Development-Challenge/tree/master/img//Communication.png)

###  Game logic

The game and the players have different states which determine the progress of the game and how the responses from players gets handled. The following are the game states:
InLobby,
AwaitingAnswers,
Busy,
Finished

The following are the player states:
Ready,
NotReady,
Start,
Answering,
Answered

When players connect, they will be put in a lobby. From the lobby players can execute the following commands:
/ready,
/unready,
/start,
/exit

When two or more players are ready, any one of the players can start the game. A countdown will begin as soon as the /start command is entered. When the game is in Busy state, it will ignore all input from the client and only processes and displays data. When the game is AwaitingAnswers state, the input from the clients will be captured, and the player's state will change to Answered.