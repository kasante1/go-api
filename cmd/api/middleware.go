package main
import (
	"fmt"
	"net"
	"net/http"
	"sync" 
	"golang.org/x/time/rate"
	"time"
	"errors" 
	"strings" 
	"github.com/kasante1/go-api/internal/validator"
	"github.com/kasante1/go-api/internal/data"
) 


func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer func() {
	
	if err := recover(); err != nil {

	w.Header().Set("Connection", "close")

	app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
	}
	}()

	next.ServeHTTP(w, r)
})

}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
		}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
		)

		go func() {
			for {
				time.Sleep(time.Minute)
			
				mu.Lock()
				
				for ip, client := range clients {

					if time.Since(client.lastSeen) > 3*time.Minute {
						delete(clients, ip)
						}

					}
					mu.Unlock()
			}
		}()
		
			
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
			next.ServeHTTP(w, r)
		}
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		
		token := headerParts[1]
		
		v := validator.New()
		
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)

		})
	}



func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	
	if !user.Activated {
		app.inactiveAccountResponse(w, r)
		return
	}
		next.ServeHTTP(w, r)
	})
		return app.requireAuthenticatedUser(fn)
	}
	
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	permissions, err := app.models.Permissions.GetAllForUser(user.ID)
	if err != nil {
	app.serverErrorResponse(w, r, err)
	return
	}

	if !permissions.Include(code) {
		app.notPermittedResponse(w, r)
		return
	}
	

	next.ServeHTTP(w, r)
	}

	return app.requireActivatedUser(fn)
}


func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin != ""{
			for i := range app.config.cors.trustedOrigins{
				if origin == app.config.cors.trustedOrigins[i]{
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""{
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}

			}
		}
		next.ServeHTTP(w, r)
	})
	}