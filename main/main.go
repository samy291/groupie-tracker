package main

import (
    "fmt"
    "html/template"
    "log"
    "database/sql"
    "errors"
    "net/http"
    "../controller"

    database "groupie-tracker/bdd"
)

func Home(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFiles("./templates/accueil.html")
    if err != nil {
        fmt.Println(err) 
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    err = t.Execute(w, nil)
    if err != nil {
        fmt.Println(err) 
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}

func loghandler(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFiles("./templates/login.html")
    if err != nil {
        log.Printf("Error parsing template: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    err = t.Execute(w, nil)
    if err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}

func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    email := r.FormValue("email")
    password := r.FormValue("password")

    db := database.InitDB()
    defer db.Close()

    isAuthorized, err := checkCredentials(db, email, password)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if isAuthorized {
        fmt.Fprintf(w, "Welcome, %s! You are now logged in.", email)
    } else {
        err := registerUser(db, email, password)
        if err != nil {
            http.Error(w, "Failed to register user", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Registration successful. Please log in again.")
    }
}


func checkCredentials(db *sql.DB, email, password string) (bool, error) {
    var id int
    var storedPassword string

    row := db.QueryRow("SELECT id, password FROM USER WHERE email = ?", email)
    err := row.Scan(&id, &storedPassword)

    if err == sql.ErrNoRows {
        return false, nil
    } else if err != nil {
        return false, err
    }

    // Vérifiez si les mots de passe correspondent
    if password == storedPassword {
        return true, nil
    }

    return false, nil
}

func registerUser(db *sql.DB, email, password string) error {
    var id int
    row := db.QueryRow("SELECT id FROM USER WHERE email = ?", email)
    err := row.Scan(&id)

    if err != sql.ErrNoRows {
        if err != nil {
            return err
        }
        return errors.New("user already exists")
    }

    _, err = db.Exec("INSERT INTO USER (email, password) VALUES (?, ?)", email, password)
    if err != nil {
        return err
    }

    return nil
}

func signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	db := database.InitDB()
	defer db.Close()

	err := registerUser(db, email, password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Registration successful. Please log in.")
}


func sign(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/sign.html")
    if err != nil {
        fmt.Println(err) 
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    err = t.Execute(w, nil)
    if err != nil {
        fmt.Println(err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}


func main() {
    http.HandleFunc("/", Home)
	http.HandleFunc("/loghandler", loghandler)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/sign", sign)
    http.HandleFunc("/login", Login)

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    fmt.Println("The server is running on port :8081")
    err := http.ListenAndServe(":8081", nil)
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    controller.TestDatabase()
}