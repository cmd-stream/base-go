# base-go
base-go provides the basis for the cmd-stream-go client and server.

The client delegates communication with the server to `ClientDelegate`, 
i.e. it sends commands and receives results with help of `ClientDelegate`. 
The server, in turn, handles connections (receives, executes commands and sends
back results) with help of `ServerDelegate`.

# Tests
Test coverage is about 93%.