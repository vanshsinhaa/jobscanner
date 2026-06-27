package common

import (
	"math/rand"
	"sync"
	"time"

	"fmt"

	"github.com/go-resty/resty/v2"
)

var (
	clientOnce sync.Once
	clientO    *resty.Client
)

func GetClient() *resty.Client {
	clientOnce.Do(func() {
		clientO = resty.New().
			SetRetryCount(3).
			// Only retry on transient failures: network errors and 429/5xx.
			// Do not retry 404/403/400 — those are permanent for this job ID.
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				sc := r.StatusCode()
				return sc == 429 || sc == 500 || sc == 502 || sc == 503 || sc == 504
			}).
			// Exponential backoff with jitter: 1–1.5s → 2–3s → 4–6s per attempt.
			// Jitter prevents goroutines from hammering the same endpoint in sync.
			SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
				attempt := r.Request.Attempt
				base := time.Duration(1<<uint(attempt-1)) * time.Second
				jitter := time.Duration(rand.Int63n(int64(base / 2)))

				fmt.Printf("retrying attempt %d, waiting %v\n", attempt, base+jitter)
				return base + jitter, nil
			})
	})
	return clientO
}

func init() {
	checkDuplicatesComapnies()
	checkAndInitWorkdayCompaniesList()
}
