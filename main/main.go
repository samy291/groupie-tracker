package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

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

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pseudo := r.FormValue("pseudo")
	email := r.FormValue("email")
	password := r.FormValue("password")

	db, err := sql.Open("sqlite3", "./groupie-tracker.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO USER (pseudo, email, password) VALUES (?, ?, ?)", pseudo, email, password)
	if err != nil {
		http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/loghandler", http.StatusSeeOther)
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

func login(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintf(w, "Login successful.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/sign", http.StatusSeeOther)
		fmt.Fprintf(w, "Invalid email or password.")
	}
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

func printUsers(db *sql.DB) {
	rows, err := db.Query("SELECT email, password FROM USER")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var email, password string
		err := rows.Scan(&email, &password)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Email:", email, "Password:", password)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func selectgame(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/selectgame.html")
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
	db := database.InitDB()
	defer db.Close()

	printUsers(db)
	http.HandleFunc("/", Home)
	http.HandleFunc("/loghandler", loghandler)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/test", selectgame)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/login", login)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("The server is running on port :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
