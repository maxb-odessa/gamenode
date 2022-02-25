package pubsub

import "sync"

// https://eli.thegreenplace.net/2020/pubsub-using-channels-in-go/

const (
	CONSUMER = 0x1000
	PRODUCER = 0x2000
)

type Topic int32

type Pubsub struct {
	sync.RWMutex
	subs map[Topic][]chan interface{}
}

func New() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[Topic][]chan interface{})
	return ps
}

func (ps *Pubsub) Subscribe(topic Topic) <-chan interface{} {
	ps.Lock()
	defer ps.Unlock()

	ch := make(chan interface{}, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *Pubsub) Publish(topic Topic, msg interface{}) {
	ps.RLock()
	defer ps.RUnlock()

	for _, ch := range ps.subs[topic] {
		go func(ch chan interface{}) {
			ch <- msg
		}(ch)
	}
}

func (ps *Pubsub) Unsubscribe(ch <-chan interface{}) {
	ps.Lock()
	defer ps.Unlock()

	for _, subs := range ps.subs {
		for idx, c := range subs {
			if c == ch {
				close(c)
				subs = removeIdx(subs, idx)
				return
			}
		}
	}

}

func removeIdx(c []chan interface{}, idx int) []chan interface{} {
	return append(c[:idx], c[idx+1:]...)
}
