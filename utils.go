package main

import (
	"os"
	"strconv"
	"net"
	"errors"
)

func (app *AppState) panicOnError(err error, action string) {
	// if err == nil {
	// 	fmt.Printf("Success: %s\n", action)
	// 	return
	// }
	if err != nil {
		app.logger.Error(action)
		os.Exit(1)
	}
}

func (app *AppState) warnOnError(err error, action string) {
	if err != nil {
		app.logger.Warn(action)
		app.logger.Warn(err)
	}
}

type ControlMessage struct {
	Cmd string
	Arg map[string]string
}


func getFreePort () (int, error) {
	for i := range 10000 {
		portNumber := i + 35565
		conn, err := net.Dial("tcp", ":" + strconv.Itoa(portNumber))
		if (err != nil) {
			return portNumber, nil
		} else {
			conn.Close()
			continue
		}

	}
	return 0, errors.New("All ports bound.")  
}

