package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestExecutableQueue(t *testing.T) {
	t.Parallel()

	type item struct {
		counter uint64
		id      uint64
	}

	ctx, cancel := context.WithCancel(context.Background())

	lock := sync.Mutex{}
	counter := uint64(0)
	items := []item{}

	q := NewExecutableQueue(func(err error) bool {
		return !common.IsContextDoneErr(err)
	}, hclog.NewNullLogger())

	go q.Execute()

	for i := 0; i < 6; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				step := j

				q.Add(func() error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Millisecond * 25):
						if id != step {
							lock.Lock()
							counter++
							newValue := counter
							items = append(items, item{counter: newValue, id: uint64(id)})
							lock.Unlock()

							fmt.Printf("from (%d, %d): %d\n", id, step, newValue)
						} else {
							return fmt.Errorf("error from %d", id)
						}
					}

					return nil
				})

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Millisecond * 5 * time.Duration(id+2)):
				}
			}
		}(i)
	}

	time.Sleep(time.Millisecond * 1500)
	cancel()
	q.Stop()

	lock.Lock()
	assert.True(t, counter > 10)
	lock.Unlock()

	time.Sleep(time.Millisecond * 500)

	lock.Lock()
	val := counter
	lock.Unlock()

	time.Sleep(time.Millisecond * 1000)

	lock.Lock()
	defer lock.Unlock()

	assert.Equal(t, val, counter)
	assert.Equal(t, val, uint64(len(items)))

	exists := map[uint64]bool{}

	for i, x := range items {
		key := x.counter<<16 + x.id

		assert.Equal(t, uint64(i+1), x.counter)
		assert.False(t, exists[key])

		exists[key] = true
	}
}
