package subpub

import (
	"sync"
)

var _ Subscription = (*subEntity)(nil)

type subEntity struct {
	once   sync.Once
	mh     MessageHandler
	name   string
	close  chan struct{}
	closed bool
	queue  chan interface{}
	id     int
}

func (s *subEntity) Unsubscribe() {
	s.once.Do(func() {
		s.close <- struct{}{}
		s.closed = true
		close(s.close)
		close(s.queue)
	})
}

func newSubEntity(subject string, mh MessageHandler) *subEntity {
	return &subEntity{
		mh:     mh,
		name:   subject,
		close:  make(chan struct{}),
		queue:  make(chan interface{}, 100),
		closed: false,
	}
}

type subscribers struct {
	mu   sync.RWMutex
	subs map[string]*partitions
}

type partitions struct {
	mu         sync.RWMutex
	partitions map[int]*subEntity
}

func (p *partitions) get(id int) *subEntity {
	p.mu.RLock()
	prtn, ok := p.partitions[id]
	p.mu.RUnlock()

	if !ok {
		return nil
	}

	return prtn
}

func (p *partitions) add(sub *subEntity) {
	p.mu.Lock()
	p.partitions[sub.id] = sub
	p.mu.Unlock()
	return
}

func (p *partitions) delete(id int) {
	p.mu.Lock()
	delete(p.partitions, id)
	p.mu.Unlock()
}

func (p *partitions) getAll() []*subEntity {
	p.mu.RLock()
	res := make([]*subEntity, 0, len(p.partitions))
	for _, sub := range p.partitions {
		res = append(res, sub)
	}
	p.mu.RUnlock()
	return res
}

func newSubscribers() *subscribers {
	return &subscribers{
		mu:   sync.RWMutex{},
		subs: make(map[string]*partitions),
	}
}

func (s *subscribers) add(entity *subEntity) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.subs[entity.name]
	if !ok {
		p = &partitions{
			mu:         sync.RWMutex{},
			partitions: make(map[int]*subEntity),
		}
	}

	entity.id = len(p.partitions)
	p.add(entity)

	s.subs[entity.name] = p
}

func (s *subscribers) get(subject string) *partitions {
	s.mu.RLock()
	p, ok := s.subs[subject]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	return p
}

func (s *subscribers) getAll() []*partitions {
	res := make([]*partitions, 0, len(s.subs))

	s.mu.RLock()
	for _, sub := range s.subs {
		res = append(res, sub)
	}
	s.mu.RUnlock()

	return res
}

func (s *subscribers) safeDelete(se *subEntity) {
	s.mu.Lock()
	p := s.subs[se.name]
	p.delete(se.id)
	s.mu.Unlock()
}
