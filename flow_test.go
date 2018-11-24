package flow

import (
	"testing"
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

	return New(awaitCommand)
}

func TestBot(t *testing.T) {
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
	data, ok := collectedData.(*UserData)
	if !ok || (data.login != "some@email.com") || (data.password != "some password") {
		t.Errorf("Wrong data: %q", collectedData)
	}
}
