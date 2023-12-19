package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/raphhawk/subs/data"
)

func (app *Config) HomePage(
	w http.ResponseWriter,
	r *http.Request,
) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	_ = app.Session.RenewToken(r.Context())

	// parse form post
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// get email and password from form post
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid Credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid Credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !validPassword {
		msg := Message{
			To:      email,
			Subject: "Failed log in attempt",
			Data:    "Invalid login attempt",
		}

		app.sendEmail(msg)

		app.Session.Put(r.Context(), "error", "Invalid Credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userId", user.ID)
	app.Session.Put(r.Context(), "user", user)
	app.Session.Put(r.Context(), "flash", "login Successful!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Config) RegisterPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	u := data.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName:  r.Form.Get("last-name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = u.Insert(u)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Unable to create user.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("http://localhost:8080/activate?email=%s", u.Email)
	signedURL := GenerateTokenFromString(url)
	app.InfoLog.Println(signedURL)

	msg := Message{
		To:       u.Email,
		Subject:  "Activate your account",
		Template: "confirmation-email",
		Data:     template.HTML(signedURL),
	}

	app.sendEmail(msg)
	app.Session.Put(r.Context(), "flash", "Confirmation email sent. Check your email.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Config) ActivateAccount(
	w http.ResponseWriter,
	r *http.Request,
) {
	url := r.RequestURI
	testURL := fmt.Sprintf("http://localhost:8080%s", url)
	if ok := VerifyToken(testURL); !ok {
		app.Session.Put(r.Context(), "error", "Invalid token.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u, err := app.Models.User.GetByEmail(r.URL.Query().Get("email"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "No User found.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u.Active = 1
	if err = u.Update(); err != nil {
		app.Session.Put(r.Context(), "error", "Unable to Update user.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.Session.Put(r.Context(), "flash", "account activated.You can now log in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Config) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {
}

func (app *Config) ChooseSubscription(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !app.Session.Exists(r.Context(), "userId") {
		app.Session.Put(r.Context(), "warning", "You must log in to see this page!.")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	dataMap := make(map[string]any)
	dataMap["plans"] = plans

	app.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: dataMap,
	})
}
