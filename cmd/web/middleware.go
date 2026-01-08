package main

import "net/http"

func (app *Application) SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}
