package subpub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	subjectName = "testSubject"
	message     = "testMessage"
)

var (
	handlerFunc = func(msg interface{}) {}
)

func TestPubSub_NewPubSub(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sp := NewSubPub()
		assert.NotNil(t, sp)
	})
}

func TestPubSub_Subscribe(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sp := NewSubPub()

		sub, err := sp.Subscribe(subjectName, handlerFunc)
		assert.NoError(t, err)
		assert.NotPanics(t, func() {
			sub.Unsubscribe()
		})
	})
	t.Run("positive", func(t *testing.T) {
		sp := &PubSub{
			mu:         sync.Mutex{},
			Subscribed: newSubscribers(),
		}

		_, err := sp.Subscribe(subjectName, handlerFunc)
		assert.NoError(t, err)

		sp.Subscribed.add(newSubEntity(subjectName, handlerFunc))
		se := sp.Subscribed.get(subjectName).get(0)

		go func() {
			se.close <- struct{}{}
			se.queue <- message
		}()
	})
}

func TestPubSub_Publish(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sp := NewSubPub()

		sub, err := sp.Subscribe(subjectName, handlerFunc)
		assert.NoError(t, err)

		err = sp.Publish(subjectName, message)
		assert.NoError(t, err)

		assert.NotPanics(t, func() {
			sub.Unsubscribe()
		})
	})
	t.Run("negative:errNoSubscribers", func(t *testing.T) {
		t.Run("case:1", func(t *testing.T) {
			sp := NewSubPub()

			sub, err := sp.Subscribe(subjectName, handlerFunc)
			assert.NoError(t, err)
			assert.NotPanics(t, func() {
				sub.Unsubscribe()
			})

			err = sp.Publish(subjectName, message)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, errNoSubscriber)
			}
		})
		t.Run("case:2", func(t *testing.T) {
			sp := NewSubPub()

			err := sp.Publish(subjectName, message)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, errNoSubscribers)
			}
		})
	})
	t.Run("negative:errTimeoutToWrite", func(t *testing.T) {
		sp := &PubSub{
			mu:         sync.Mutex{},
			Subscribed: newSubscribers(),
		}

		sp.Subscribed.add(newSubEntity(subjectName, handlerFunc))
		se := sp.Subscribed.get(subjectName).get(0)

		wg := sync.WaitGroup{}
		wg.Add(100)
		for i := 0; i <= 100; i++ {
			go func() {
				defer wg.Done()
				se.queue <- message
			}()
		}
		wg.Wait()

		err := sp.Publish(subjectName, message)
		if assert.Error(t, err) {
			assert.ErrorIs(t, err, errTimeoutToWrite)
		}
	})
}

func TestPubSub_Close(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sp := &PubSub{
			mu:         sync.Mutex{},
			Subscribed: newSubscribers(),
		}

		_, err := sp.Subscribe(subjectName, handlerFunc)
		assert.NoError(t, err)

		err = sp.Publish(subjectName, message)
		assert.NoError(t, err)

		ctx := context.Background()

		err = sp.Close(ctx)
		assert.NoError(t, err)

		// should wait while subscribe collect data and close chan
		time.Sleep(100 * time.Millisecond)

		assert.Len(t, sp.Subscribed.get(subjectName).partitions, 0)
	})
	t.Run("negative:context is canceled", func(t *testing.T) {
		sp := &PubSub{
			Subscribed: newSubscribers(),
		}

		_, err := sp.Subscribe(subjectName, handlerFunc)
		assert.NoError(t, err)

		err = sp.Publish(subjectName, message)
		assert.NoError(t, err)

		ctx := context.Background()
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)

		cancel()

		err = sp.Close(ctx)
		if assert.Error(t, err) {
			assert.ErrorIs(t, err, context.Canceled)
		}

		err = sp.Publish(subjectName, message)
		assert.NoError(t, err)
	})
	t.Run("negative:context deadline exceed", func(t *testing.T) {
		sp := &PubSub{
			Subscribed: newSubscribers(),
		}

		for i := 0; i <= 100; i++ {
			_, err := sp.Subscribe(subjectName, handlerFunc)
			assert.NoError(t, err)
		}

		err := sp.Publish(subjectName, message)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		err = sp.Close(ctx)
		if err != nil {
			assert.ErrorIs(t, err, context.Canceled)
		}
	})
	t.Run("negative:context timeout while closing", func(t *testing.T) {
		sp := &PubSub{
			Subscribed: newSubscribers(),
		}

		for range 100 {
			_, err := sp.Subscribe(subjectName, handlerFunc)
			assert.NoError(t, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := sp.Close(ctx)
		if assert.Error(t, err) {
			assert.ErrorIs(t, err, context.DeadlineExceeded)
		}
	})
}
