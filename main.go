package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var tpl *template.Template
var userData = make(map[string]User)

// Session Store
var sessionStore = make(map[string]string)
var sessionStoreLock sync.RWMutex

type User struct {
	Name     string
	Email    string
	Password string
}
type PageData struct {
	Message string
}

// lsof -i :8080
// kill -9

func main() {
	tpl, _ = template.ParseGlob("template/*.html")

	// Create a new  (router)
	mux := http.NewServeMux()
	// It creates a new router that allows you to define multiple routes for your web server.

	// Define routes

	mux.HandleFunc("/signup", signupmethod)
	mux.HandleFunc("/login", loginfunc)
	mux.HandleFunc("/loginpost", loginmethod)
	mux.HandleFunc("/logout", logoutfunc)
	mux.HandleFunc("/clear", clearUser)
	mux.HandleFunc("/", indexHandle)
	mux.HandleFunc("/home", homefunc)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", authMiddleware(mux))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(" Middleware executed for:", r.URL.Path)

		publicRoutes := map[string]bool{
			"/signup":     true,
			"/signuppost": true,
			"/login":      true,
			"/loginpost":  true,
		}

		if _, isPublic := publicRoutes[r.URL.Path]; isPublic {
			fmt.Println("Public route allowed:", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Check if session cookie exists
		cookie, err := r.Cookie("sessionID")
		if err != nil {
			fmt.Println(" No session cookie found. Redirecting to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session ID exists in session store
		sessionStoreLock.RLock() // Using a lock for thread safety
		userEmail, sessionExists := sessionStore[cookie.Value]
		sessionStoreLock.RUnlock()

		if !sessionExists || userEmail == "" {
			fmt.Println(" Invalid session. Redirecting to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		fmt.Println(" User authenticated! Access granted to:", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func homefunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	tpl.ExecuteTemplate(w, "index.html", nil)
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func signupmethod(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			n := PageData{Message: "Invalid: All fields are required"}
			tpl.ExecuteTemplate(w, "signup.html", n)
			return
		}

		if _, exists := userData[email]; exists {
			n := PageData{Message: "User already exists"}
			tpl.ExecuteTemplate(w, "signup.html", n)
			return
		}

		userData[email] = User{Name: name, Email: email, Password: password}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "signup.html", PageData{Message: ""})
}

func loginfunc(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginmethod(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("emailLogin")
	password := r.FormValue("passwordLogin")

	user, exists := userData[email]

	if email == "" || password == "" {
		n := PageData{Message: "Invalid"}
		tpl.ExecuteTemplate(w, "login.html", n)
		return
	}

	if !exists || user.Password != password {
		n := PageData{Message: "Invalid credentials"}
		tpl.ExecuteTemplate(w, "login.html", n)
		return
	}

	// Generate a unique session ID
	sessionID := strconv.Itoa(rand.Intn(1000000)) + strconv.FormatInt(time.Now().Unix(), 10)
	sessionStore[sessionID] = email

	// generate cookie
	sessionCookie := &http.Cookie{
		Name:   "sessionID",
		Value:  sessionID,
		MaxAge: 300,
		Path:   "/",
	}
	http.SetCookie(w, sessionCookie)
	fmt.Println(" Login successful, session created!")

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func logoutfunc(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("sessionID")
	if err == nil {
		delete(sessionStore, cookie.Value)
	}

	expiredCookie := &http.Cookie{
		Name:   "sessionID",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	}
	http.SetCookie(w, expiredCookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func clearUser(w http.ResponseWriter, r *http.Request) {
	userData = make(map[string]User)
	sessionStore = make(map[string]string) // Clear all sessions
	tpl.ExecuteTemplate(w, "signup.html", nil)
	fmt.Println(" All user data and sessions cleared!")
}
