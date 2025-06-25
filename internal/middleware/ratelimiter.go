package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type UserLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

func NewUserRateLimiter(r rate.Limit, b int) *UserLimiter {
	return &UserLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (u *UserLimiter) getLimiter(key string) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter, exists := u.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(u.r, u.b)
		u.limiters[key] = limiter
	}
	return limiter
}

func RateLimitMiddleware(limiter *UserLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			if userID, ok := GetUserID(r.Context()); ok {
				key = "user:" + fmt.Sprint(userID)
			} else {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					ip = r.RemoteAddr
				}
				key = "ip:" + ip
			}
			if !limiter.getLimiter(key).Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
