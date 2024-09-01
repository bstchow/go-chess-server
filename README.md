# Go Chess Server

The go-chess-server project is a backend server application designed to manage chess games and provide a platform for multiplayer matches. Built using the Go programming language, the server supports essential functionalities for hosting online chess matches, including player authentication, game state management, move validation, and real-time game updates

![pixlr-image-generator-6b8608fb-840f-4d55-94a9-a3b4b4b9f8d8](https://github.com/yelaco/go-chess-server/assets/100106895/8c43dd20-cc83-4da8-a6f1-0c9c62692e66)

## Features

**Player Authentication**
- Login/Registration: Users can create accounts and log in
- Session management: The server handles user sessions to maintain login states and manage active connections

**Game Management**
- Matchmaking: Players can enter matching queue and wait for another player to create a match. If a player leave the match, he/she can come back later by rejoin the match.
- Game state: The server maintains the state of ongoing games, tracking each move and updating the board accordingly.
- Data persistence: After a game ended, its information is saved to database, ensuring that game states are preserved and can be retrieved later for user's analysis purposes.
  
**Move Handling**
- Move validation: The server validates each move to ensure they are legal according to chess rules.
- The server checks for conditions like check, checkmate, and stalemate after each move to determine the game's status.
  
**Real-Time Updates**
- WebSocket integration: The server uses WebSockets to provide real-time updates to connected clients, ensuring players see the latest game state without needing to refresh.

## How to run

### Server

**Docker**

** The Docker setup has not been tested after changes from forked repo**

###

## API

### REST

 [References](https://documenter.getpostman.com/view/30874401/2sA3duEsiX)
 
- ```POST /api/users```: To register user
- ```POST /api/login```: To log in to the server
- ```GET /api/sessions```: Retrieve match records played by user
- ```GET /api/sessions/{sessionid}```: Retrieve single match record based on ID

### WebSocket

After login, user can now join a match by sending matching request
```json
{
    "action": "matching",
    "data": {
        "playerId": "12345"
    }
}
```

If the ```action``` and ```data``` is valid, server pushes that user into the matching queue. When a match happens, the two connections are forwarded to game management module, where a game instance will be initialized and binded with the player pair. Then, a message is sent back to the user to notify about the match.
```json
{
    "type": "matched",
    "session_id": "1232524",
    "game_state": {
        "status": "ACTIVE",
        "board": [[]]
        "is_white_turn": true,
    },
    "player_state": {
        "is_white_side": true
    }
}
```

On the contrary, if there are any errors in the process or the matching request is timeout, the server replies with
- Error (Note that this error json is universal for all the error response to users)
```json
{
    "type": "error",
    "error": "<err_msg>",
}
```

- Timeout
```json
{
    "type": "timeout",
    "error": "<err_msg>",
}
```

In a match, users can send move request with 
```json
{
    "action": "move",
    "data": {
        "playerId": "12345",
        "sessionId": "1719199808062498696",
        "move": "e2-e4"
    }
}
```

And get resonses as 
```json
{
    "type": "session",
    "game_state": {
        "status": "STALEMATE",
        "board": [[]]
        "is_white_turn": false,
    }
}
```

After the game reaches end state, the server notifies both players and close their connections.
