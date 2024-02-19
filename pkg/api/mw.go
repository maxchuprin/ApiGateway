// MiddleWare для поддержки сквозного id запроса, логгирования и установки заголовков ответа
package api

import (
	"ApiGateway/pkg/obj"
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// HeadersMiddleware устанавливает заголовки ответа сервера.
func (api *API) HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Manual-Header", " ")
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware читает из запроса requestID или генерирует его и записывает в контекст
func (api *API) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idParam := r.URL.Query().Get("requestID")
		var id int
		var err error
		if idParam != "" {
			id, err = strconv.Atoi(idParam)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			id = 1000000 + rand.Intn(10000000)
		}

		ctx := context.WithValue(r.Context(), obj.ContextKey("requestID"), id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// структура и метод для логгирования http кода ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (c *loggingResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.ResponseWriter.WriteHeader(statusCode)
}

// миддлваре для логгирования ответов
func (api *API) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logWR := &loggingResponseWriter{ResponseWriter: w}
		// Call the next handler in the chain with custom ResponseWriter that saves http code
		next.ServeHTTP(logWR, r)
		// After the request handler is called
		log.Printf("at %v from %v request id %v was proccesed with http-code %v", time.Now(), r.RemoteAddr, r.Context().Value(obj.ContextKey("requestID")), logWR.statusCode)
	})
}
