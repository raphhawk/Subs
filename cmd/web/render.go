package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/raphhawk/subs/data"
)

var pathToTemplate = "cmd/web/templates"

type TemplateData struct {
	StringMap             map[string]string
	IntMap                map[string]int
	FloatMap              map[string]float64
	Data                  map[string]any
	Flash, Warning, Error string
	Authenticated         bool
	Now                   time.Time
	User                  *data.User
}

func (app *Config) isAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userId")
}

func (app *Config) AddDefaultData(
	td *TemplateData,
	r *http.Request,
) *TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.isAuthenticated(r) {
		td.Authenticated = true
		user, ok := app.Session.Get(r.Context(), "user").(data.User)
		if !ok {
			app.ErrorLog.Println("cant get user for session")
		} else {
			td.User = &user
		}
	}
	td.Now = time.Now()
	return td
}

func (app *Config) render(
	w http.ResponseWriter,
	r *http.Request,
	t string,
	td *TemplateData,
) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", pathToTemplate),
		fmt.Sprintf("%s/header.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/navbar.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/alerts.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/footer.partial.gohtml", pathToTemplate),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf(
		"%s/%s",
		pathToTemplate,
		t,
	))

	templateSlice = append(templateSlice, partials...)
	if td == nil {
		td = &TemplateData{}
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, app.AddDefaultData(td, r)); err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
