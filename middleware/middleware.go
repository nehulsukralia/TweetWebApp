package middleware

import(
	"net/http"
	"davyWybiralWebApp/session"
)

//middleware
func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if user is logged in by checking the values in the session otherwise redirect to the login page
		session, _ := session.Store.Get(r, "session")
		_, ok := session.Values["user_id"]
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return
		}
		handler.ServeHTTP(w, r)
	}
}