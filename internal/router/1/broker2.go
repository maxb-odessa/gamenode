/*
  originated from here: https://stackoverflow.com/questions/36417199/how-to-broadcast-message-using-channel
*/
package broadcaster

import (
	"fmt"
	"sync"
)

type Broadcaster struct {
	sync.Mutex
	clients map[int64]chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[int64]chan struct{}),
	}
}

func (br *Broadcaster) Subscribe(id int64) (<-chan struct{}, error) {
	defer br.Unlock()
	br.Lock()
	s := make(chan struct{}, 1)

	if _, ok := br.clients[id]; ok {
		return nil, fmt.Errorf("signal %d already exist", id)
	}

	bir.clients[id] = s

	return br.clients[id], nil
}

func (br *Broadcaster) Unsubscribe(id int64) {
	defer br.Unlock()
	br.Lock()
	if _, ok := br.clients[id]; ok {
		close(br.clients[id])
	}

	delete(br.clients, id)
}

func (br *Broadcaster) broadcast() {
	defer br.Unlock()
	br.Lock()
	for k := range br.clients {
		if len(br.clients[k]) == 0 {
			br.clients[k] <- struct{}{}
		}
	}
}

/*
type testClient struct {
	name     string
	signal   <-chan struct{}
	signalID int64
	brd      *Broadcaster
}

func (c *testClient) doWork() {
	i := 0
	for range c.signal {
		fmt.Println(c.name, "do work", i)
		if i > 2 {
			c.brd.Unsubscribe(c.signalID)
			fmt.Println(c.name, "unsubscribed")
		}
		i++
	}
	fmt.Println(c.name, "done")
}

func main() {
	var err error
	brd := NewBroadcaster()

	clients := make([]*testClient, 0)

	for i := 0; i < 3; i++ {
		c := &testClient{
			name:     fmt.Sprint("client:", i),
			signalID: time.Now().UnixNano() + int64(i), // +int64(i) for play.golang.org
			brd:      brd,
		}
		c.signal, err = brd.Subscribe(c.signalID)
		if err != nil {
			log.Fatal(err)
		}

		clients = append(clients, c)
	}

	for i := 0; i < len(clients); i++ {
		go clients[i].doWork()
	}

	for i := 0; i < 6; i++ {
		brd.broadcast()
		time.Sleep(time.Second)
	}
}

*/
