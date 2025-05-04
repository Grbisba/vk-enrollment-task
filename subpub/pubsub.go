package subpub

import (
	"context"
	"sync"
	"time"
)

var (
	_ SubPub = (*PubSub)(nil)
)

type PubSub struct {
	mu         sync.Mutex
	Subscribed *subscribers
}

// MessageHandler is a callback function that processes messages delivered to subscribers.
type MessageHandler func(msg interface{})

// Subscribe should subscribe to name and consume messages.
func (ps *PubSub) Subscribe(subject string, mh MessageHandler) (Subscription, error) {
	se := newSubEntity(subject, mh)

	ps.mu.Lock()
	ps.Subscribed.add(se)
	ps.mu.Unlock()

	go func() {
		for {
			select {
			case v := <-se.queue:
				se.mh(v)
			case <-se.close:
				for v := range se.queue {
					se.mh(v)
				}
				se.Unsubscribe()
				ps.Subscribed.safeDelete(se)
				return
			case <-time.After(3 * time.Minute):
				return
			}
		}
	}()

	return se, nil
}

// Publish should send messages to the name.
func (ps *PubSub) Publish(subject string, msg interface{}) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	p := ps.Subscribed.get(subject)
	if p == nil || len(p.partitions) == 0 {
		return errNoSubscribers
	}

	err := ps.publishData(p, msg)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PubSub) publishData(p *partitions, msg interface{}) error {
	var err error
	pLen := len(p.partitions)

	if pLen == 0 {
		return errNoSubscribers
	}

	wg := sync.WaitGroup{}
	wg.Add(pLen)

	for i, cp := range p.partitions {
		go func(id int, cp *subEntity) {
			defer wg.Done()
			if cp == nil || cp.closed {
				err = errNoSubscriber
				return
			}

			select {
			case cp.queue <- msg:
			case <-time.After(time.Second):
				err = errTimeoutToWrite
				return
			}
		}(i, cp)
	}
	wg.Wait()

	return err

}

// Close will shutdown pub-sub system.
// May be blocked by data delivery until the context is canceled.
func (ps *PubSub) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return ctx.Err()
	}

	done := make(chan struct{})
	cancel := make(chan struct{})

	go func() {
		ps.mu.Lock()
		for _, topic := range ps.Subscribed.getAll() {
			for _, se := range topic.getAll() {
				select {
				case _ = <-cancel:
					return
				default:
					se.Unsubscribe()
				}
			}
		}
		ps.mu.Unlock()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		go func() {
			cancel <- struct{}{}
		}()

		return ctx.Err()
	case <-done:
		return nil
	}
}

func NewSubPub() SubPub {
	s := &PubSub{
		Subscribed: newSubscribers(),
	}

	return s
}
