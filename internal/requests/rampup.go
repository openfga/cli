package requests

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RampUpAPIRequests gradually increases the rate of API requests from minRPS to maxRPS over rampUpPeriodInSec seconds.
// After reaching maxRPS, it maintains this rate until all requests are sent or the context is canceled.
// Parameters:
// - ctx: The context to control cancellation.
// - minRPS: The initial requests per second.
// - maxRPS: The maximum requests per second.
// - rampUpPeriodInSec: The period over which to ramp up the request rate.
// - maxInFlight: The maximum number of concurrent requests, used to protect the client and the server from being overwhelmed.
// - requests: A slice of functions representing the API requests to be made.
func RampUpAPIRequests(ctx context.Context, minRPS, maxRPS, rampUpPeriodInSec, maxInFlight int, requests []func() error) error {
	rpsIncrement := float64(maxRPS-minRPS) / float64(rampUpPeriodInSec)
	limiter := rate.NewLimiter(rate.Limit(minRPS), 1)
	semaphore := make(chan struct{}, maxInFlight)
	var wg sync.WaitGroup

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	requestIndex := 0

	for step := 0; step <= rampUpPeriodInSec; step++ {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case <-ticker.C:
			if err := limiter.Wait(ctx); err != nil {
				wg.Wait()
				return err
			}

			for i := 0; i < int(limiter.Limit()); i++ {
				if requestIndex >= len(requests) {
					wg.Wait()
					return nil
				}

				semaphore <- struct{}{}
				wg.Add(1)

				go func(req func() error) {
					defer wg.Done()
					defer func() { <-semaphore }()
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
			wg.Wait()
			return ctx.Err()
		case <-ticker.C:
			if err := limiter.Wait(ctx); err != nil {
				wg.Wait()
				return err
			}

			for i := 0; i < int(limiter.Limit()); i++ {
				if requestIndex >= len(requests) {
					wg.Wait()
					return nil
				}

				semaphore <- struct{}{}
				wg.Add(1)

				go func(req func() error) {
					defer wg.Done()
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
