package main

import (
	"database/sql"
	"fmt"

	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	database "groupie-tracker/bdd"
	groupieWebsocket "groupie-tracker/groupieWebsocket"
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

func AddRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	createdBy := "default"
	maxPlayer := r.FormValue("max_player")

	db, err := sql.Open("sqlite3", "./groupie-tracker.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO ROOMS (name, created_by, max_player) VALUES (?, ?, ?)", name, createdBy, maxPlayer)
	if err != nil {
		http.Error(w, "Failed to insert room into database", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to retrieve the result of the database operation", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		fmt.Println("No room was added to the database.")
	} else {
		fmt.Println("A room was successfully added to the database.")
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
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

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
		http.ServeFile(w, r, "templates/selectgame.html")

	} else {
		// http.Redirect(w, r, "/loghandler", http.StatusSeeOther)
		http.ServeFile(w, r, "templates/login.html")
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

func printRooms(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM ROOMS")
	if err != nil {
		return err
	}
	defer rows.Close()

	var id, created_by, max_player, id_game int
	var name, code string
	fmt.Println("ID | Created By | Max Players | Name | Game ID | Code")
	for rows.Next() {
		err = rows.Scan(&id, &created_by, &max_player, &name, &id_game, &code)
		if err != nil {
			return err
		}
		fmt.Printf("%d | %d | %d | %s | %d | %s\n", id, created_by, max_player, name, id_game, code)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
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

func profile(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/profile.html")
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

func createroom(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/createroom.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/create.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func join(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/room.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func createblind(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/create/createblind.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func main() {
	db := database.InitDB()
	defer db.Close()

	printRooms(db)
	printUsers(db)

	http.HandleFunc("/", Home)
	http.HandleFunc("/loghandler", loghandler)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/test", selectgame)
	http.HandleFunc("/Signup", Signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/profile", profile)
	http.HandleFunc("/createroom", createroom)
	http.HandleFunc("/create", create)
	http.HandleFunc("/join", join)
	http.HandleFunc("/createblind", func(w http.ResponseWriter, r *http.Request) {
		createblind(db, w, r)
	})
	http.HandleFunc("/addroom", AddRoom)

	http.HandleFunc("/websocket", groupieWebsocket.WebsocketHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("The server is running on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
