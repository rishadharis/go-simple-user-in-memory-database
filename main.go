package main

import (
	"html/template"
	"net/http"

	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var tpl *template.Template
var dbSession = map[string]string{}
var dbUser = map[string]user{}

type user struct {
	Name     string
	Username string
	Email    string
	Password []byte
}

func indexPage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// req.Cookie("session")
	tpl.ExecuteTemplate(res, "index.gohtml", nil)
}

func loginPage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/dashboard", http.StatusSeeOther)
		return
	}
	var data string
	failed := req.FormValue("failed")
	status := req.FormValue("status")
	if failed == "true" {
		data = "Username atau password salah"
	}
	if status != "" {
		switch status {
		case "prohibited":
			data = "Anda dilarang mengakses halaman ini sebelum login"
		case "logout":
			data = "Anda berhasil logout"
		default:
			data = "Login terlebih dahulu"
		}
	}
	tpl.ExecuteTemplate(res, "login.gohtml", data)
}

func login(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if data, ok := dbUser[username]; ok {
		err := bcrypt.CompareHashAndPassword(data.Password, []byte(password))
		if err != nil {
			http.Redirect(res, req, "/login?failed=true", http.StatusSeeOther)
			return
		}
		uuid := uuid.NewV4()
		http.SetCookie(res, &http.Cookie{
			Name:  "session",
			Value: uuid.String(),
		})
		dbSession[uuid.String()] = username
		http.Redirect(res, req, "/dashboard", http.StatusSeeOther)

	}

}

func logout(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cook, err := req.Cookie("session")
	if err != nil {
		http.Redirect(res, req, "/login?status=prohibited", http.StatusSeeOther)
		return
	}
	delete(dbSession, cook.Value)
	cook.MaxAge = -1
	http.SetCookie(res, cook)
	http.Redirect(res, req, "/login?status=logout", http.StatusSeeOther)
}

func dashboardPage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/login?status=prohibited", http.StatusSeeOther)
		return
	}
	data := getData(req)
	tpl.ExecuteTemplate(res, "dashboard.gohtml", data)

}

func registerPage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/dashboard", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(res, "register.gohtml", nil)
}

func register(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// req.Cookie("session")
	username := req.PostFormValue("username")
	pass, err := bcrypt.GenerateFromPassword([]byte(req.PostFormValue("password")), bcrypt.MinCost)
	if err != nil {
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := user{
		Name:     req.PostFormValue("name"),
		Email:    req.PostFormValue("email"),
		Username: username,
		Password: pass,
	}
	dbUser[username] = data
	// tpl.ExecuteTemplate(res, "test.gohtml", req.PostForm["email"][0])
	uuid := uuid.NewV4()
	http.SetCookie(res, &http.Cookie{
		Name:  "session",
		Value: uuid.String(),
	})

	dbSession[uuid.String()] = username
	http.Redirect(res, req, "/dashboard", http.StatusSeeOther)
	// tpl.ExecuteTemplate(res, "register.gohtml", nil)
}

func init() {
	tpl = template.Must(template.ParseGlob("contents/*.gohtml"))
	tpl = template.Must(tpl.ParseGlob("contents/template/*.gohtml"))
}

func main() {
	mux := httprouter.New()
	mux.GET("/", indexPage)

	mux.GET("/login", loginPage)
	mux.POST("/login", login)

	mux.GET("/register", registerPage)
	mux.POST("/register", register)

	mux.GET("/dashboard", dashboardPage)

	mux.GET("/logout", logout)

	mux.GET("/assets/*filepath", FileServerHandler("./assets"))
	// http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.ListenAndServe(":8080", mux)
}

// FileServerHandler creates a custom handler for serving static files using http.FileServer
func FileServerHandler(root string) httprouter.Handle {
	fileServer := http.FileServer(http.Dir(root))
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Extract the file path from the URL parameters
		// filePath := ps.ByName("filepath")
		// Use http.StripPrefix to serve the correct file
		http.StripPrefix("/assets", fileServer).ServeHTTP(w, r)
	}
}
