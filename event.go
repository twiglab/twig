package twig

import "container/list"

type EventFunc func(*Event)

type Emitter interface {
	Emit(string, *Event)
}

type Receivre interface {
	On(string, *Event)
}

type Event struct {
	Sync int
	Body interface{}
}

type Messager interface {
	Emitter
	On(string, Receivre)
}

type events map[string]list.List

type eBus struct {
	eventList events
}

func newEventBus() *eBus {
	return &eBus{
		eventList: make(events),
	}
}

func (e *eBus) On(event string, r Receivre) {
	topic, ok := e.eventList[event]

	if !ok {
		topic = list.List{}
	}

	topic.PushBack(r)
	e.eventList[event] = topic
}

func (e *eBus) Emit(event string, msg *Event) {
	go func() {
		if topic, ok := e.eventList[event]; ok {
			for el := topic.Front(); el != nil; el = el.Next() {
				r := el.Value.(Receivre)
				r.On(event, msg)
			}
		} else {
		}
	}()
}
