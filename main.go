package main

import (
	"davyWybiralWebApp/models"
	"davyWybiralWebApp/routes"
	"davyWybiralWebApp/utils"
	"net/http"
)

func main() {
	// setting up storage
	models.Init()

	// load templates
	utils.LoadTemplate("templates/*.html")

	// router
	r := routes.NewRouter()

	// start server
	http.ListenAndServe(":8080", r)
}
