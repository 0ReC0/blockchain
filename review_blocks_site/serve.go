package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	fmt.Println("🌍 Сервер запущен на http://localhost:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
