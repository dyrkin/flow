package main

import (
	"fmt"

	. "github.com/dyrkin/flow"
)

//Human data storage
type UserData struct {
	login    string
	password string
}

func (u *UserData) String() string {
	return fmt.Sprintf("UserData{login: %q, password: %q}", u.login, u.password)
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

//human emulator
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

	return New(askRegister)
}

func main() {
	humanChan := make(chan string)
	botChan := make(chan string)

	//bot conversation flow
	bot := newBot(humanChan)
	//human conversation flow
	human := newHuman(botChan)

	go func() {
		for {
			select {
			//receive massage from human and redirect it to bot
			case toBot := <-botChan:
				bot.Send(toBot)
			//receive massage from bot and redirect it to human
			case toHuman := <-humanChan:
				human.Send(toHuman)
			}
		}
	}()

	human.Start()
	//lock until the end of bot flow
	collectedData := <-bot.Start()
	fmt.Println(collectedData)
}
