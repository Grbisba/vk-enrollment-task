package subpub

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe_newSubEntity(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := newSubEntity(subjectName, handlerFunc)
		assert.NotNil(t, e)
	})
}

func TestSubscribe_Unsubscribe(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := newSubEntity(subjectName, handlerFunc)
		assert.NotNil(t, e)

		go func() {
			for {
				select {
				case <-e.close:
					return
				}
			}
		}()

		e.Unsubscribe()
	})
	t.Run("positive:double unsubscribe", func(t *testing.T) {
		e := newSubEntity(subjectName, handlerFunc)
		assert.NotNil(t, e)

		go func() {
			for {
				select {
				case <-e.close:
					return
				}
			}
		}()

		e.Unsubscribe()
		assert.NotPanics(t, func() {
			e.Unsubscribe()
		})
	})
}

func TestSubscribers_addSubscription(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := newSubscribers()
		subject := "123"

		se := newSubEntity(subject, func(msg interface{}) {
			fmt.Printf("%s recieved message: %s\n", subject, msg)
		})

		s.add(se)

		assert.Equal(t, 1, len(s.get(subject).partitions))
		assert.Equal(t, s.get("123"), &partitions{
			mu:         sync.RWMutex{},
			partitions: map[uuid.UUID]*subEntity{se.id: se},
		})

		se2 := newSubEntity(subject, func(msg interface{}) {
			fmt.Printf("%s recieved message: %s\n", subject, msg)
		})

		s.add(se2)

		assert.Equal(t, 2, len(s.get(subject).partitions))
		t.Log(s.get("123"))
		assert.Equal(t, s.get("123"), &partitions{
			mu:         sync.RWMutex{},
			partitions: map[uuid.UUID]*subEntity{se.id: se, se2.id: se2},
		})
	})
}

func TestSubscribers_getSubscription(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := newSubscribers()
		subject := "123"

		se := newSubEntity(subject, func(msg interface{}) {
			fmt.Printf("%s recieved message: %s\n", subject, msg)
		})

		s.add(se)

		assert.Equal(t, 1, len(s.get(subject).partitions))
		assert.Equal(t, s.get("123"), &partitions{
			mu:         sync.RWMutex{},
			partitions: map[uuid.UUID]*subEntity{se.id: se},
		})

		se2 := newSubEntity(subject, func(msg interface{}) {
			fmt.Printf("%s recieved message: %s\n", subject, msg)
		})

		s.add(se2)

		assert.Equal(t, 2, len(s.get(subject).partitions))
		assert.Equal(t, s.get("123"), &partitions{
			mu:         sync.RWMutex{},
			partitions: map[uuid.UUID]*subEntity{se.id: se, se2.id: se2},
		})

		subs := s.get("123").partitions
		assert.Equal(t, 2, len(subs))
		assert.Equal(t, subs[se.id], se)
		assert.Equal(t, subs[se.id], se)
	})
}

func TestSubscribers_removeSubscription(t *testing.T) {}
