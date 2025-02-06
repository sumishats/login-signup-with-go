package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var tpl *template.Template
var userData = make(map[string]User)

// ✅ Session Store
var sessionStore = make(map[string]string) // sessionID -> email

type User struct {
	Name     string
	Email    string
	Password string
}

func main() {
	tpl, _ = template.ParseGlob("template/*.html")

	// Create a new ServeMux (router)
	mux := http.NewServeMux()

	// Define routes
	mux.HandleFunc("/signup", handlefuncSignup)
	mux.HandleFunc("/signuppost", signupmethod)
	mux.HandleFunc("/login", loginfunc)
	mux.HandleFunc("/loginpost", loginmethod)
	mux.HandleFunc("/logout", logoutfunc)
	mux.HandleFunc("/clear", clearUser)
	mux.HandleFunc("/", indexHandle)
	mux.HandleFunc("/home", homefunc)

	// Wrap the entire router with authentication middleware
	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", authMiddleware(mux))
}

// ✅ Middleware for Authentication
func authMiddleware(next http.Handler) http.Handler {
	fmt.Println("iam middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🚀 Middleware executed for:", r.URL.Path)

		// Define public routes (allowed without login)
		publicRoutes := map[string]bool{
			"/signup":     true,
			"/signuppost": true,
			"/login":      true,
			"/loginpost":  true,
		}

		// Allow public routes without authentication
		if _, isPublic := publicRoutes[r.URL.Path]; isPublic {
			fmt.Println("✅ Public route allowed:", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Check for authentication session
		cookie, err := r.Cookie("sessionID")
		if err != nil || sessionStore[cookie.Value] == "" {
			fmt.Println("❌ Unauthorized access! Redirecting to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		fmt.Println("✅ User authenticated! Access granted to:", r.URL.Path)
		next.ServeHTTP(w, r) // Continue to the requested handler
	})
}

// ✅ Home Page (Requires Authentication)
func homefunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	tpl.ExecuteTemplate(w, "index.html", nil)
}

// ✅ Index Handler (Redirects to Home)
func indexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🏠 Redirecting to home...")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// ✅ Signup Page (Public Route)
func handlefuncSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📝 Signup page accessed")
	err := tpl.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		http.Error(w, "Signup page not found: "+err.Error(), http.StatusInternalServerError)
	}
}

// ✅ Signup Processing (Public Route)
func signupmethod(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔑 Signup process started")

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		tpl.ExecuteTemplate(w, "signup.html", "All fields are required")
		return
	}

	if _, exists := userData[email]; exists {
		tpl.ExecuteTemplate(w, "signup.html", "User already exists")
		return
	}

	userData[email] = User{Name: name, Email: email, Password: password}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ✅ Login Page (Public Route)
func loginfunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔓 Login page accessed")
	tpl.ExecuteTemplate(w, "login.html", nil)
}

// ✅ Login Processing (Public Route)
func loginmethod(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔑 Login process started")

	email := r.FormValue("emailLogin")
	password := r.FormValue("passwordLogin")

	user, exists := userData[email]

	if email == "" || password == "" {
		tpl.ExecuteTemplate(w, "login.html", "Both fields are required")
		return
	}

	if !exists || user.Password != password {
		tpl.ExecuteTemplate(w, "login.html", "Invalid credentials")
		return
	}

	// ✅ Generate a unique session ID
	sessionID := strconv.Itoa(rand.Intn(1000000)) + strconv.FormatInt(time.Now().Unix(), 10)
	sessionStore[sessionID] = email // Store session in session map

	// ✅ Set Authentication Cookie
	sessionCookie := &http.Cookie{
		Name:   "sessionID",
		Value:  sessionID,
		MaxAge: 300, // 5 minutes
		Path:   "/",
	}
	http.SetCookie(w, sessionCookie)
	fmt.Println("✅ Login successful, session created!")

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// ✅ Logout Function (Destroys Cookie and Session)
func logoutfunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🚪 Logging out...")

	// Get session ID from cookie
	cookie, err := r.Cookie("sessionID")
	if err == nil {
		delete(sessionStore, cookie.Value) // Remove session from store
	}

	// Expire the authentication cookie
	expiredCookie := &http.Cookie{
		Name:   "sessionID",
		Value:  "",
		MaxAge: -1, // Expire the cookie immediately
		Path:   "/",
	}
	http.SetCookie(w, expiredCookie)
	fmt.Println("✅ Logout successful!")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ✅ Clears all user data (For Testing)
func clearUser(w http.ResponseWriter, r *http.Request) {
	userData = make(map[string]User)
	sessionStore = make(map[string]string) // Clear all sessions
	tpl.ExecuteTemplate(w, "signup.html", nil)
	fmt.Println("🔄 All user data and sessions cleared!")
}
