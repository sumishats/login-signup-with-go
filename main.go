package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
)

var tpl *template.Template
var userData = make(map[string]User)

type PageData struct {
	EmailInvalid string
	PassInvalid  string
}
type User struct {
	Name     string
	Email    string
	Password string
}

var RandumNumber = rand.Intn(200)

func main() {

	tpl, _ = template.ParseGlob("template/*.html")

	//epozhano route verunne athil ullil ulla func wrk avanum aa funcil must ayi response request parameter venam
	http.HandleFunc("/signup", handlefuncSignup)
	http.HandleFunc("/signuppost", signupmethod)
	http.HandleFunc("/login", loginfunc)
	http.HandleFunc("/loginpost", loginmethod)
	http.HandleFunc("/home", homefunc)
	http.HandleFunc("/logout", logoutfunc)
	http.HandleFunc("/", indexHandle)
	http.HandleFunc("/clear", clearUser)

	http.ListenAndServe(":8080", nil)

}

func clearUser(w http.ResponseWriter, r *http.Request) {

	userData = make(map[string]User)

	tpl.ExecuteTemplate(w, "signup.html", nil)
	fmt.Println(userData)
}

func homefunc(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("logincookie")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	tpl.ExecuteTemplate(w, "index.html", nil)

}
func indexHandle(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("logincookie")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/home", http.StatusSeeOther)

}

func handlefuncSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("entered the singup ")

	cookie, err := r.Cookie("logincookie")
	if err == nil && cookie.Value != "" {
		// Optionally redirect or display message for logged-in users
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	err = tpl.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		http.Error(w, "Signup page is not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
func signupmethod(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" {
		tpl.ExecuteTemplate(w, "signup.html", "Email is nil")
		return
	}
	if password == "" {
		tpl.ExecuteTemplate(w, "signup.html", "Password is nil")
		return
	}
	if name == "" {
		tpl.ExecuteTemplate(w, "signup.html", "Name is nil")
		return
	}

	if _, ok := userData[email]; ok {
		tpl.ExecuteTemplate(w, "signup.html", "User already found")
		return
	}

	userData[email] = User{Name: name, Email: email, Password: password}
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

func loginfunc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	cookie, err := r.Cookie("logincookie")
	if err == nil && cookie.Value == strconv.Itoa(RandumNumber) {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.html", nil)
}
func loginmethod(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("logincookie")
	if err != nil {
		fmt.Println(err)
	} else if cookie.Value != "" {
		http.Redirect(w, r, "loginpost", http.StatusSeeOther)
		return
	}
	email := r.FormValue("emailLogin")
	password := r.FormValue("passwordLogin")

	user, ok := userData[email]

	if email == "" {
		n := PageData{EmailInvalid: "email not found"}
		tpl.ExecuteTemplate(w, "login.html", n)
		fmt.Println("email is empty")
		return
	} else if password == "" {
		n := PageData{PassInvalid: "password not found"}
		tpl.ExecuteTemplate(w, "login.html", n)
		return
	}

	if ok && password != user.Password {
		n := PageData{PassInvalid: "Invalid Credentials"}
		tpl.ExecuteTemplate(w, "login.html", n)
		return
	}

	if ok && password == user.Password {
		fmt.Println("reached cookie creatiion")
		CookirForLogin := &http.Cookie{}
		CookirForLogin.Name = "logincookie"
		CookirForLogin.Value = strconv.Itoa(RandumNumber)
		CookirForLogin.MaxAge = 300
		CookirForLogin.Path = "/"
		http.SetCookie(w, CookirForLogin)
		tpl.ExecuteTemplate(w, "index.html", CookirForLogin.Value)

	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}

func logoutfunc(w http.ResponseWriter, r *http.Request) {
	cookielogout := http.Cookie{}
	cookielogout.Name = "logincookie"
	cookielogout.Value = ""
	cookielogout.MaxAge = -1
	cookielogout.Path = "/"
	http.SetCookie(w, &cookielogout)

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.Redirect(w, r, "/login", http.StatusSeeOther)

	cookie, err := r.Cookie("logincookie")
	if err != nil {
		fmt.Println(err)

	} else if cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}
