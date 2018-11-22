package conversation

type Data interface{}
type Message interface{}

type ReplyFunction func(event *Event) *NextStep
type askFunction func(data Data)

type NextStep struct {
	step *Step
	data Data
}

type Event struct {
	Message Message
	Data    Data
}

type Step struct {
	askFn   askFunction
	replyFn ReplyFunction
}

type ask struct {
	askFn askFunction
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
	return func(event *Event) *NextStep {
		panic("Oooops. Something went wrong. Define your own default ReplyFunction")
	}
}

type Conversation struct {
	askChan chan *Step

	replyChan chan Message
}

func StartWithData(initialStep *Step, initialData Data) *Conversation {
	conversation := &Conversation{askChan: make(chan *Step), replyChan: make(chan Message)}
	processor := func() {
		var currentData = initialData
		for {
			step := <-conversation.askChan
			if step.askFn != nil {
				step.askFn(currentData)
			}
			reply := <-conversation.replyChan
			if step.replyFn == nil {
				return
			}
			nextStep := step.replyFn(&Event{reply, currentData})
			if nextStep.data != nil {
				currentData = nextStep.data
			}
			if nextStep.step == nil {
				return
			}
			go func() {
				conversation.askChan <- nextStep.step
			}()
		}
	}
	go processor()
	conversation.askChan <- initialStep
	return conversation
}

func Start(initialStep *Step) *Conversation {
	return StartWithData(initialStep, nil)
}

func (conversation *Conversation) Send(message Message) {
	go func() {
		conversation.replyChan <- message
	}()
}
