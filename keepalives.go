package main

func handlePing(_ ControlMessage, routing Routing, app *AppState) {
	response := PingResponse{Status: "ok", Name: "Pong!"}
	rpcReply(app, routing, response)
}


