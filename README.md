# Flow - Design and control conversation flow

## Overview

Model the structured conversation flow between the user and a chatbot, or between whatever you want.

## A Simple Example

```go
import (
	"fmt"
	. "github.com/dyrkin/flow"
)

//Human data storage
type UserData struct {
	login    string
	password string
}

//bot emulator
func newBot(humanChan chan string) *Flow {
	var awaitCommand *Step
	var askEmail *Step
	var askPassword *Step

	//wait for command from human
	awaitCommand =
		OnReply(func(msg Message, data Data) *NextStep {
			switch msg {
			case "register":
				return Goto(askEmail).Using(&UserData{})
			case "quit":
				return End()
			}
			return DefaultHandler()(msg, data)
		})

	//ask human for email
	askEmail =
		Ask(func(data Data) {
			humanChan <- "please send your email"
		}).OnReply(func(msg Message, data Data) *NextStep {
			email := msg.(string)
			humanData := data.(*UserData)
			humanData.login = email
			return Goto(askPassword)
		})

	//ask human for password
	askPassword =
		Ask(func(data Data) {
			humanChan <- "please send your password"
		}).OnReply(func(msg Message, data Data) *NextStep {
			password := msg.(string)
			humanData := data.(*UserData)
			humanData.password = password
			return End().Using(humanData)
		})

	return NewFlow(awaitCommand)
}

func main() {
	humanChan := make(chan string)

	//bot conversation flow
	bot := newBot(humanChan)

	go func() {
		messages := []string{"some@email.com", "some password"}
		index := 0
		for {
			select {
			case <-humanChan:
				bot.Send(messages[index])
				index++
			}
		}
	}()

	dataChan := bot.Start()

	bot.Send("register")

	collectedData := <-dataChan
	fmt.Println(collectedData)
}
```

The basic strategy is to define Steps and specifying initial step while instantiating the Flow:

* `NewFlow(<initial step>)` creates a flow with the initial step specified.
* `Start()` starts the flow and returns a `chan` where the collected data will be sent after end of the flow.
* `Ask(<data fn>)` executes immediately after `Start()` is called.
* `OnReply(<message fn>)` executes when a data from the user is received.
* `<message fn>` must return next step or end the flow

The code will produce the following output:

> UserData{login: "some@email.com", password: "some password"}

Full working example can be found there: [example/conversation.go](https://github.com/dyrkin/flow/blob/master/example/conversation.go)