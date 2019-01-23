package twig

import "container/list"

type Event struct {
	Sync int
	Body interface{}
	Kind int
}

type EventHandler interface {
	On(string, *Event)
}

type EventReactor interface {
	Emit(string, *Event)
	Register(string, EventHandler)
}

type Eventer interface {
	EventReactor(EventReactor)
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
				r.On(event, msg)
			}
		}
	}()
}

func (b *ebox) Register(topic string, eh EventHandler) {
	hs, ok := b.eventList[topic]

	if !ok {
		hs = list.List{}
	}

	hs.PushBack(eh)
	b.eventList[topic] = hs
}

// EventEmitter 事件发送接口
type EventEmitter interface {
	Emit(string, *Event)
}
