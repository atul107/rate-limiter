package main

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]map[string]int
	// client   *redis.Client
	limits   map[string]int
	interval time.Duration
	mutex    sync.Mutex
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: redisAddr,
	// })
	return &RateLimiter{
		requests: make(map[string]map[string]int),
		// client:   rdb,
		limits:   make(map[string]int),
		interval: interval,
	}
}

func (rl *RateLimiter) AddRule(endpoint string, limit int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.limits[endpoint] = limit
}

func (rl *RateLimiter) MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Path
		clientKey := getClientKey(r)

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		if _, ok := rl.requests[endpoint]; !ok {
			rl.requests[endpoint] = make(map[string]int)
		}

		// count, err := rl.client.Get(r.Context(), getRedisKey(endpoint, clientKey)).Int()
		// if err != nil && err != redis.Nil {
		// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
		// 	return
		// }

		if rl.requests[endpoint][clientKey] >= rl.limits[endpoint] {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		rl.requests[endpoint][clientKey]++
		// err = rl.client.Incr(r.Context(), getRedisKey(endpoint, clientKey)).Err()
		// if err != nil {
		// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
		// 	return
		// }

		// rl.client.Expire(r.Context(), getRedisKey(endpoint, clientKey), rl.interval)
		go func() {
			<-time.After(rl.interval)
			rl.mutex.Lock()
			defer rl.mutex.Unlock()
			rl.requests[endpoint][clientKey]--
		}()
		next.ServeHTTP(w, r)
	})
}

// func getRedisKey(endpoint string, client string) string {
// 	return endpoint + ":" + client
// }

func getClientKey(r *http.Request) string {
	return r.Header.Get("client-key")
}

func main() {
	rl := NewRateLimiter(1 * time.Minute)
	rl.AddRule("/api/endpoint", 10)

	mux := http.NewServeMux()
	mux.Handle("/api/endpoint", rl.MiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request successful."))
	})))

	http.ListenAndServe(":8080", mux)
}
