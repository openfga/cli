package requests

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// RampUpAPIRequests gradually increases the rate of API requests from minRPS to maxRPS over rampUpPeriodInMin seconds.
// After reaching maxRPS, it maintains this rate until all requests are sent or the context is canceled.
// Parameters:
// - ctx: The context to control cancellation.
// - minRPS: The initial requests per second.
// - maxRPS: The maximum requests per second.
// - rampUpPeriod: The period over which to ramp up the request rate.
// - rampupPeriodDuration: The unit of each ramp-up period.
// - maxInFlight: The maximum number of concurrent requests, used to protect the client and the server.
// - requests: A slice of functions representing the API requests to be made.
func RampUpAPIRequests( //nolint:cyclop
	ctx context.Context,
	minRPS, maxRPS, rampUpPeriod int, rampupPeriodDuration time.Duration, maxInFlight int,
	requests []func() error,
) error {
	var (
		rpsIncrement = float64(maxRPS-minRPS) / float64(rampUpPeriod)
		limiter      = rate.NewLimiter(rate.Limit(minRPS), 1)
		semaphore    = make(chan struct{}, maxInFlight)
		waitGroup    sync.WaitGroup
		ticker       = time.NewTicker(rampupPeriodDuration)
		requestIndex int32
	)

	// if the ramp up period is 0, go to max rps directly
	if rampUpPeriod == 0 {
		minRPS = maxRPS
	}

	defer ticker.Stop()

	requestsLen := int32(len(requests)) //nolint:gosec

	worker := func(req func() error) {
		defer waitGroup.Done()
		defer func() { <-semaphore }()

		if req != nil {
			_ = req()
		}
	}

	for step := 0; step <= rampUpPeriod; step++ {
		select {
		case <-ctx.Done():
			waitGroup.Wait()

			return ctx.Err() //nolint:wrapcheck
		case <-ticker.C:
			if err := limiter.Wait(ctx); err != nil {
				waitGroup.Wait()

				return err //nolint:wrapcheck
			}

			for i := 0; i < int(limiter.Limit()); i++ { //nolint:intrange
				idx := atomic.AddInt32(&requestIndex, 1) - 1
				if idx >= requestsLen {
					waitGroup.Wait()

					return nil
				}

				req := requests[idx]

				semaphore <- struct{}{}

				waitGroup.Add(1)

				go worker(req)
			}

			newRPS := rate.Limit(float64(minRPS) + rpsIncrement*float64(step))
			limiter.SetLimit(newRPS)
		}
	}

	limiter.SetLimit(rate.Limit(maxRPS))

	for {
		select {
		case <-ctx.Done():
			waitGroup.Wait()

			return ctx.Err() //nolint:wrapcheck
		case <-ticker.C:
			if err := limiter.Wait(ctx); err != nil {
				waitGroup.Wait()

				return err //nolint:wrapcheck
			}

			for i := 0; i < int(limiter.Limit()); i++ { //nolint:intrange
				idx := atomic.AddInt32(&requestIndex, 1) - 1
				if idx >= requestsLen {
					waitGroup.Wait()

					return nil
				}

				req := requests[idx]

				semaphore <- struct{}{}

				waitGroup.Add(1)

				go worker(req)
			}
		}
	}
}
