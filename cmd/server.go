// сервер ApiGateway
package main

import (
	"APIGateway/pkg/api"
	"log"
	"net/http"
)

func main() {

	// Создание объекта API
	api := api.New()

	// Запуск сетевой службы и HTTP-сервера на всех локальных IP-адресах на порту 8080.
	err := http.ListenAndServe(":8080", api.Router())
	if err != nil {
		log.Fatal(err)
	}

}
