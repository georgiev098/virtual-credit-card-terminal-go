package main

import "net/http"

func (app *Application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", nil); err != nil {
		app.ErrorLog.Println(err)
	}
}
