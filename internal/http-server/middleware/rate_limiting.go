package middleware

import (
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

func RateLimiting() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		type client struct {
			limiter  *rate.Limiter
			lastSeen time.Time
		}
		var (
			mu      sync.Mutex
			clients = make(map[string]*client)
		)
		go func() {
			for {
				time.Sleep(time.Minute)

				mu.Lock()
				for ip, c := range clients {
					if time.Since(c.lastSeen) > 3*time.Minute {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			}
		}()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(1, 1)}
			}

			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				http.Error(w, "too many request", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
