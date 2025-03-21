package requests

import (
	"context"
	"fmt"
	"sync"
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
func RampUpAPIRequests( //nolint:gocognit,cyclop
	ctx context.Context,
	minRPS, maxRPS, rampUpPeriod int, rampupPeriodDuration time.Duration, maxInFlight int,
	requests []func() error,
) error {
	rpsIncrement := float64(maxRPS-minRPS) / float64(rampUpPeriod)
	limiter := rate.NewLimiter(rate.Limit(minRPS), 1)
	semaphore := make(chan struct{}, maxInFlight)

	var waitGroup sync.WaitGroup

	ticker := time.NewTicker(rampupPeriodDuration)
	defer ticker.Stop()

	requestIndex := 0
	requestsLen := len(requests)

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
				if requestIndex >= requestsLen {
					waitGroup.Wait()

					return nil
				}

				semaphore <- struct{}{}

				waitGroup.Add(1)

				go func(req func() error) {
					defer waitGroup.Done()
					defer func() { <-semaphore }()

					if req == nil {
						fmt.Printf("Error: request function is nil, request %d out of %d\n", requestIndex, requestsLen)

						return
					}

					if err := req(); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}(requests[requestIndex])

				requestIndex++
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
				if requestIndex >= len(requests) {
					waitGroup.Wait()

					return nil
				}

				semaphore <- struct{}{}

				waitGroup.Add(1)

				go func(req func() error) {
					defer waitGroup.Done()
					defer func() { <-semaphore }()

					if err := req(); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}(requests[requestIndex])

				requestIndex++
			}
		}
	}
}
