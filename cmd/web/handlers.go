package main

import "net/http"

func (app *Application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	app.InfoLog.Println("Hit the handler")
}
