package handlers

import (
	"fmt"
	"github.com/boyter/pincer/common"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

func IpRestrictorHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetIP(r)

		if strings.Contains(ip, ":") {
			ip = strings.TrimSpace(strings.Split(ip, ":")[0])
		}

		_ipMapMutex.Lock()

		result, ok := _ipMap[ip]
		if ok {
			// continue to ban to a certain level beyond which they just have to wait a long time
			if result < 1_000 {
				result++
				_ipMap[ip] = result
			}
			_ipMapMutex.Unlock()

			if result >= RateLimit {
				log.Info().Str(common.UniqueCode, "203d577e").Msg(fmt.Sprintf("%s has too many requests throttling", ip))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Header().Set("Content-Type", jsonContentType)
				w.Header().Set("Retry-After", fmt.Sprintf("%d", result))

				// lel
				_, _ = fmt.Fprint(w, `{"url":"https://en.wikipedia.org/wiki/Exponential_backoff"}`)
				return
			}

		} else {
			_ipMap[ip] = 1
			_ipMapMutex.Unlock()
		}

		next.ServeHTTP(w, r)
	}
}

func CORSHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", jsonContentType)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func CORSRateLimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return CORSHandler(IpRestrictorHandler(next))
}

func EmptyHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}
