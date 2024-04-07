package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFiles("./templates/accueil.html")
    if err != nil {
        log.Fatal(err)
    }
    t.Execute(w, nil) 
}

func main() {
    http.HandleFunc("/", Home)

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    fmt.Println("The server is running on port :8080")
    http.ListenAndServe(":8080", nil)
    
}