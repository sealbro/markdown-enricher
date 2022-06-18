package batching

import (
	"time"
)

type BatchItem interface {
}

// Process https://elliotchance.medium.com/batch-a-channel-by-size-or-time-in-go-92fa3098f65
func Process[TItem BatchItem](values <-chan TItem, maxItems int, maxTimeout time.Duration) chan []TItem {
	batches := make(chan []TItem)

	go func() {
		defer close(batches)

		for keepGoing := true; keepGoing; {
			var batch []TItem
			expire := time.After(maxTimeout)
			for {
				select {
				case value, ok := <-values:
					if !ok {
						keepGoing = false
						goto done
					}

					batch = append(batch, value)
					if len(batch) == maxItems {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if len(batch) > 0 {
				batches <- batch
			}
		}
	}()

	return batches
}
