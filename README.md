# base-go
base-go serves as the foundation for building the cmd-stream-go client and 
server.

The client delegates communication with the server to `ClientDelegate`, which 
handles sending commands and receiving results. Similarly, the server uses 
`ServerDelegate` to manage connections, execute commands, and send back results.

# Tests
Test coverage is about 93%.