package main

import (
	"fmt"

	. "github.com/dyrkin/flow"
)

type UserData struct {
	login    string
	password string
}

//bot emulator
func newBot(humanChan chan string) *Flow {
	var awaitCommand *Step
	var askEmail *Step
	var askPassword *Step

	//wait for command from user
	awaitCommand =
		OnReply(func(msg Message, data Data) *NextStep {
			switch msg {
			case "register":
				return Goto(askEmail).Using(&UserData{})
			}
			return DefaultHandler()(msg, data)
		})

	//ask email
	askEmail =
		Ask(func(data Data) {
			humanChan <- "please send your email"
		}).OnReply(func(msg Message, data Data) *NextStep {
			email := msg.(string)
			userData := data.(*UserData)
			userData.login = email
			return Goto(askPassword)
		})

	//ask password
	askPassword =
		Ask(func(data Data) {
			humanChan <- "please send your password"
		}).OnReply(func(msg Message, data Data) *NextStep {
			password := msg.(string)
			userData := data.(*UserData)
			userData.password = password
			return End().Using(userData)
		})

	return Start(awaitCommand)
}

//user emulator
func newHuman(botChan chan string) *Flow {
	var askRegister *Step
	var sendEmail *Step
	var sendPassword *Step

	//send regiter command to bot and process response
	askRegister =
		Ask(func(data Data) {
			botChan <- "register"
		}).OnReply(func(msg Message, data Data) *NextStep {
			switch msg {
			case "please send your email":
				return Goto(sendEmail)
			}
			return DefaultHandler()(msg, data)
		})

	//send email to bot and process response
	sendEmail =
		Ask(func(data Data) {
			botChan <- "some@email.com"
		}).OnReply(func(msg Message, data Data) *NextStep {
			switch msg {
			case "please send your password":
				return Goto(sendPassword)
			}
			return DefaultHandler()(msg, data)
		})

	//just send a password to the bot and stop the flow
	sendPassword =
		Ask(func(data Data) {
			botChan <- "some password"
		})

	return Start(askRegister)
}

func main() {
	humanChan := make(chan string)
	botChan := make(chan string)

	bot := newBot(humanChan)
	human := newHuman(botChan)

	go func() {
		for {
			select {
			//receive massage from user and redirect it to bot
			case toBot := <-botChan:
				bot.Send(toBot)
			//receive massage from bot and redirect it to user
			case toHuman := <-humanChan:
				human.Send(toHuman)

			}
		}
	}()

	//lock until the end of human flow
	human.DataSync()
	//lock until the end of bot flow
	completeData := bot.DataSync()
	fmt.Printf("Complete data: %q\n", completeData)
}
