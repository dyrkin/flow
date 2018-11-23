package flow

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
	askChan      chan *Step
	replyChan    chan Message
	completeData chan Data
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
		panic("Oooops. Something went wrong. Define your own default ReplyFunction")
	}
}

func StartWithData(initialStep *Step, initialData Data) *Flow {
	flow := &Flow{
		askChan:      make(chan *Step),
		replyChan:    make(chan Message),
		completeData: make(chan Data),
	}
	processor := func() {
		var currentData = initialData
		for {
			step := <-flow.askChan
			if step.askFn != nil {
				step.askFn(currentData)
			}
			if step.replyFn == nil {
				go func() {
					flow.completeData <- currentData
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
					flow.completeData <- currentData
				}()
				return
			}
			go func() {
				flow.askChan <- nextStep.step
			}()
		}
	}
	go processor()
	flow.askChan <- initialStep
	return flow
}

func Start(initialStep *Step) *Flow {
	return StartWithData(initialStep, nil)
}

func (flow *Flow) Send(message Message) {
	go func() {
		flow.replyChan <- message
	}()
}

func (flow *Flow) DataSync() Data {
	return <-flow.DataAsync()
}

func (flow *Flow) DataAsync() chan Data {
	return flow.completeData
}
