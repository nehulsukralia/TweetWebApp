package routes

import (
	"davyWybiralWebApp/middleware"
	"davyWybiralWebApp/models"
	"davyWybiralWebApp/session"
	"davyWybiralWebApp/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	// handler setup
	r := mux.NewRouter()
	r.HandleFunc("/", middleware.AuthRequired(IndexGetHandler)).Methods("GET")
	r.HandleFunc("/", middleware.AuthRequired(IndexPostHandler)).Methods("POST")
	r.HandleFunc("/login", LoginGetHandler).Methods("GET")
	r.HandleFunc("/login", LoginPostHandler).Methods("POST")
	r.HandleFunc("/logout", LogoutGetHandler).Methods("GET")
	r.HandleFunc("/register", RegisterGetHandler).Methods("GET")
	r.HandleFunc("/register", RegisterPostHandler).Methods("POST")
	r.HandleFunc("/{username}", middleware.AuthRequired(UserGetHandler)).Methods("GET")


	// host static files(css etc)
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	return r
}

func IndexGetHandler(w http.ResponseWriter, r *http.Request) {
	// get the all updates from the storage and show them on the page
	updates, err := models.GetAllUpdates()
	if err != nil { // if redis is down or something
		utils.InternalServerError(w)
		return
	}
	utils.ExecuteTemplate(w, "index.html", struct {
		Title string
		Updates []*models.Update
		DisplayForm bool
	} {
		Title: "All Updates",
		Updates: updates,
		DisplayForm: true,
	})
}

func IndexPostHandler(w http.ResponseWriter, r *http.Request) {
	// get userID from the session which is required in PostUpdate function
	session, _ := session.Store.Get(r, "session")
	untypedUserID := session.Values["user_id"] // untypedUserID because session hold empty interface objects
	userID, ok := untypedUserID.(int64)        // so we are doing type assertion since userID is untyped till now
	if !ok {
		utils.InternalServerError(w)
		return
	}

	// get the user update from the form
	r.ParseForm()
	body := r.PostForm.Get("update")

	// store the update to the storage
	err := models.PostUpdate(userID, body)
	if err != nil {
		utils.InternalServerError(w)
		return
	}

	// Redirect the user to the same page
	http.Redirect(w, r, "/", 302)

}

func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	// get userID of the logged in user from the session which is required in PostUpdate function
	session, _ := session.Store.Get(r, "session")
	untypedUserID := session.Values["user_id"] // untypedUserID because session hold empty interface objects
	loggedInuserID, ok := untypedUserID.(int64)        // so we are doing type assertion since userID is untyped till now
	if !ok {
		utils.InternalServerError(w)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	// get userID by username
	user, err := models.GetUserByUsername(username)
	if err != nil { // if redis is down or something
		utils.InternalServerError(w)
		return
	}

	userID, err := user.GetID()
	if err != nil { // if redis is down or something
		utils.InternalServerError(w)
		return
	}
	
	// get the user updates from the storage and show them on the page
	updates, err := models.GetUpdates(userID)
	
	if err != nil { // if redis is down or something
		utils.InternalServerError(w)
		return
	}
	utils.ExecuteTemplate(w, "index.html", struct {
		Title string
		Updates []*models.Update
		DisplayForm bool
	} {
		Title: username,
		Updates: updates,
		DisplayForm: loggedInuserID == userID, //display submit form only if the user is the one currently logged in
	})
}

func LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "login.html", nil)
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	// get username and password from form
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	user, err :=models.AuthenticateUser(username, password)
	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			utils.ExecuteTemplate(w, "login.html", "unknown user")
		case models.ErrInvalidLogin:
			utils.ExecuteTemplate(w, "login.html", "invalid login")
		default:
			utils.InternalServerError(w)
		}
		return
	}

	// after successful login store userID in the session
	userID, err := user.GetID()
	if err != nil {
		utils.InternalServerError(w)
		return
	}
	session, _ := session.Store.Get(r, "session")
	session.Values["user_id"] = userID
	session.Save(r, w)

	// after successful login redirect user to the index page
	http.Redirect(w, r, "/", 302)
}

func LogoutGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/login", 302)
}

func RegisterGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "register.html", nil)
}

func RegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	// get username and password from form
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	err := models.RegisterUser(username, password)
	if err == models.ErrUsernameTaken {
		utils.ExecuteTemplate(w, "register.html", "username taken")
		return
	} else if err != nil { //if redis is down
		utils.InternalServerError(w)
		return
	}

	// redirect user to login page after registration
	http.Redirect(w, r, "/login", 302)
}
