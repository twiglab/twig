package twig

import (
	"container/list"
)

type EventReceiver interface {
	On(EventRegister)
}

type EventHandlerFunc func(string, *Event)

func (eh EventHandlerFunc) OnEvent(topic string, ev *Event) {
	eh(topic, ev)
}

type EventHandler interface {
	Handler(string, *Event)
}

type Event struct {
	Sync int
	Body interface{}
	Kind int
}

type EventEmitter interface {
	Emit(string, *Event)
}

type EventRegister interface {
	On(string, EventHandler) EventRegister
}

type Messager interface {
	EventEmitter
	EventRegister
}

type events map[string]list.List
type ebox struct {
	eventList events
}

func newbox() *ebox {
	return &ebox{
		eventList: make(events),
	}
}

func (b *ebox) Emit(event string, msg *Event) {
	go func() {
		if topic, ok := b.eventList[event]; ok {
			for el := topic.Front(); el != nil; el = el.Next() {
				r := el.Value.(EventHandler)
				r.Handler(event, msg)
			}
		}
	}()
}

func (b *ebox) On(topic string, eh EventHandler) EventRegister {
	hs, ok := b.eventList[topic]

	if !ok {
		hs = list.List{}
	}

	hs.PushBack(eh)
	b.eventList[topic] = hs

	return b
}
