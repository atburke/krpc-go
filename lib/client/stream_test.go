package client

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/atburke/krpc-go/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestStreamManager(t *testing.T) {
	streamCounts := []int{0, 1, 2, 10}
	input := []string{"this", "is", "the", "test", "input"}
	for _, numStreams := range streamCounts {
		numStreams := numStreams
		t.Run(fmt.Sprintf("%v stream(s) listening", numStreams), func(t *testing.T) {
			sm := newStreamManager()
			streamData := map[int][]string{}
			mu := sync.Mutex{}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			for i := 0; i < numStreams; i++ {
				i := i
				stream := sm.newStream()
				go func() {
					for {
						select {
						case data := <-stream.C:
							mu.Lock()
							streamData[i] = append(streamData[i], string(data))
							mu.Unlock()
						case <-ctx.Done():
							return
						}
					}
				}()
			}

			// HACK: Stream values will likely be dropped if they're sent too
			// quickly. Fortunately, in a real stream a) there will be a delay
			// between updates anyway and b) it's ok to drop a few values.
			for _, s := range input {
				time.Sleep(10 * time.Millisecond)
				sm.write([]byte(s))
			}

			require.Eventually(t, func() bool {
				mu.Lock()
				defer mu.Unlock()
				for _, data := range streamData {
					if !utils.SlicesEqual(data, input) {
						return false
					}
				}
				return true
			}, 1000*time.Millisecond, 100*time.Millisecond, "map contents: %v", streamData)
		})
	}
}
