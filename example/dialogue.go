package main

import (
	"fmt"

	. "github.com/dyrkin/conversation"
)

type UserData struct {
	login    string
	password string
}

//bot emulator
func newBot(humanChan chan string) *Conversation {
	var awaitCommand *Step
	var askEmail *Step
	var askPassword *Step

	//wait for command from user
	awaitCommand = OnReply(
		func(event *Event) *NextStep {
			switch event.Message {
			case "register":
				return Goto(askEmail).Using(&UserData{})
			}
			return DefaultHandler()(event)
		})

	//ask email
	askEmail = Ask(
		func(data Data) {
			humanChan <- "please send your email"
		}).OnReply(
		func(event *Event) *NextStep {
			email := event.Message.(string)
			userData := event.Data.(*UserData)
			userData.login = email
			return Goto(askPassword)
		})

	//ask password
	askPassword = Ask(func(data Data) {
		humanChan <- "please send your password"
	}).OnReply(func(event *Event) *NextStep {
		password := event.Message.(string)
		userData := event.Data.(*UserData)
		userData.password = password
		fmt.Printf("Complete data: %q", userData)
		return End().Using(userData)
	})

	return Start(awaitCommand)
}

//user emulator
func newHuman(botChan chan string) *Conversation {
	var askRegister *Step
	var sendEmail *Step
	var sendPassword *Step

	//send regiter command to bot and process response
	askRegister = Ask(
		func(data Data) {
			botChan <- "register"
		}).OnReply(
		func(event *Event) *NextStep {
			switch event.Message {
			case "please send your email":
				return Goto(sendEmail)
			}

			return DefaultHandler()(event)
		})

	//send email to bot and process response
	sendEmail = Ask(func(data Data) {
		botChan <- "some@email.com"
	}).OnReply(func(event *Event) *NextStep {
		switch event.Message {
		case "please send your password":
			return Goto(sendPassword)
		}
		return DefaultHandler()(event)
	})

	//send password to bot and stop the flow
	sendPassword = Ask(func(data Data) {
		botChan <- "some password"
	}).OnReply(func(event *Event) *NextStep {
		return End()
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
			case toHuman := <-humanChan:
				//receive massage from bot and redirect it to user
				human.Send(toHuman)

			}
		}
	}()

	var in string
	fmt.Scanln(&in)
}
