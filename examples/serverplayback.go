// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package examples

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	. "weavelab.xyz/goesl"
)

var welcomeFile = "%s/media/welcome.wav"

func main() {

	defer func() {
		if r := recover(); r != nil {
			Error("Recovered in f", r)
		}
	}()

	// Boost it as much as it can go ...
	runtime.GOMAXPROCS(runtime.NumCPU())

	wd, err := os.Getwd()

	if err != nil {
		Error("Error while attempt to get WD: %s", wd)
		os.Exit(1)
	}

	welcomeFile = fmt.Sprintf(welcomeFile, wd)

	if s, err := NewOutboundServer(":8084"); err != nil {
		Error("Got error while starting Freeswitch outbound server: %s", err)
	} else {
		go handle(s)
		s.Start()
	}

}

// handle - Running under goroutine here to explain how to handle playback ( play to the caller )
func handle(s *OutboundServer) {
	ctx := context.Background()

	for {

		select {

		case conn := <-s.Conns:
			Notice("New incomming connection: %v", conn)

			if err := conn.Connect(ctx); err != nil {
				Error("Got error while accepting connection: %s", err)
				break
			}

			answer, err := conn.ExecuteAnswer(ctx, "", false)

			if err != nil {
				Error("Got error while executing answer: %s", err)
				break
			}

			Debug("Answer Message: %s", answer)
			Debug("Caller UUID: %s", answer.GetHeader("Caller-Unique-Id"))

			cUUID := answer.GetCallUUID()

			if sm, err := conn.Execute(ctx, "playback", welcomeFile, true); err != nil {
				Error("Got error while executing playback: %s", err)
				break
			} else {
				Debug("Playback Message: %s", sm)
			}

			if hm, err := conn.ExecuteHangup(ctx, cUUID, "", false); err != nil {
				Error("Got error while executing hangup: %s", err)
				break
			} else {
				Debug("Hangup Message: %s", hm)
			}

			go func() {
				for {
					msg, err := conn.ReadMessage(ctx)

					if err != nil {

						// If it contains EOF, we really dont care...
						if !strings.Contains(err.Error(), "EOF") {
							Error("Error while reading Freeswitch message: %s", err)
						}
						break
					}

					Debug("%s", msg)
				}
			}()

		default:
			// YabbaDabbaDooooo!
			//Flintstones. Meet the Flintstones. They're the modern stone age family. From the town of Bedrock,
			// They're a page right out of history. La la,lalalalala la :D
		}
	}

}
