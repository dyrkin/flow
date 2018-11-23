package flow

import "fmt"

type Data interface{}
type Message interface{}

type ReplyFunction func(msg Message, data Data) *NextStep
type askFunction func(data Data)

type NextStep struct {
	step *Step
	data Data
}

type Step struct {
	askFn   askFunction
	replyFn ReplyFunction
}

type ask struct {
	askFn askFunction
}

type Flow struct {
	askChan     chan *Step
	replyChan   chan Message
	initialStep *Step
	initialData Data
}

func Ask(askFn askFunction) *Step {
	return &Step{askFn: askFn}
}

func (step *Step) OnReply(replyFn ReplyFunction) *Step {
	step.replyFn = replyFn
	return step
}

func OnReply(replyFn ReplyFunction) *Step {
	return &Step{nil, replyFn}
}

func Goto(step *Step) *NextStep {
	return &NextStep{step: step}
}

func End() *NextStep {
	return &NextStep{}
}

func (nextStep *NextStep) Using(data Data) *NextStep {
	nextStep.data = data
	return nextStep
}

func DefaultHandler() ReplyFunction {
	return func(msg Message, data Data) *NextStep {
		errorMassage := fmt.Sprintf("Oooops. Something went wrong. Define your own default ReplyFunction.\nMessage: %q. Data: %q", msg, data)
		panic(errorMassage)
	}
}

func NewWithData(initialStep *Step, initialData Data) *Flow {
	flow := &Flow{
		askChan:     make(chan *Step, 1),
		replyChan:   make(chan Message, 1),
		initialStep: initialStep,
		initialData: initialData,
	}
	return flow
}

func New(initialStep *Step) *Flow {
	return NewWithData(initialStep, nil)
}

func (flow *Flow) Send(message Message) {
	go func() {
		flow.replyChan <- message
	}()
}

func (flow *Flow) Start() chan Data {
	collectedData := make(chan Data, 1)
	processor := func() {
		var currentData = flow.initialData
		for {
			step := <-flow.askChan
			if step.askFn != nil {
				step.askFn(currentData)
			}
			if step.replyFn == nil {
				go func() {
					collectedData <- currentData
				}()
				return
			}
			reply := <-flow.replyChan
			nextStep := step.replyFn(reply, currentData)
			if nextStep.data != nil {
				currentData = nextStep.data
			}
			if nextStep.step == nil {
				go func() {
					collectedData <- currentData
				}()
				return
			}
			go func() {
				flow.askChan <- nextStep.step
			}()
		}
	}
	go processor()
	flow.askChan <- flow.initialStep
	return collectedData
}
