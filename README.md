# base-go
base-go provides the basis for the cmd-stream client and server. 

The client delegates communication with the server to the `ClientDelegate`, 
i.e. it sends commands and receives results with help of the `ClientDelegate`. 
The server, in turn, handles connections (receives, executes commands, and sends
back results) with help of the `ServerDelegate`.

# Tests
Test coverage is about 93%.