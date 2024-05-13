package main

import (
	"database/sql"
	"fmt"
	database "groupie-tracker/bdd"
	blindtest "groupie-tracker/game"
	"groupie-tracker/groupieWebsocket"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Room struct {
	ID        *int
	CreatedBy *string
	MaxPlayer *int
	Name      *string
	Mode      *string
	GameID    *int
}

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
	modeCookie, err := r.Cookie("mode")
	if err != nil {
		http.Error(w, "Mode not found", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	mode := modeCookie.Value

	if mode == "" {
		http.Error(w, "Mode is required", http.StatusBadRequest)
		return
	}

	// Get the pseudo from the cookie
	pseudo, err := r.Cookie("pseudo")
	if err != nil {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}
	createdBy := pseudo.Value

	maxPlayer := r.FormValue("max_player")
	db, err := sql.Open("sqlite3", "./groupie-tracker.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	result, err := db.Exec("INSERT INTO ROOMS (name, created_by, max_player, mode) VALUES (?, ?, ?, ?)", name, createdBy, maxPlayer, mode) // Include the mode in the database insertion
	if err != nil {
		fmt.Println(err) // Print the error message to the console
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
		http.ServeFile(w, r, "templates/selectgame.html")
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

	// Check if the username or email already exists
	var exists int
	err = db.QueryRow("SELECT count(*) FROM USER WHERE pseudo = ? OR email = ?", pseudo, email).Scan(&exists)
	if err != nil {
		http.Error(w, "Failed to verify uniqueness of user", http.StatusInternalServerError)
		return
	}
	if exists != 0 {
		http.Error(w, "Username or email already exists", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO USER (pseudo, email, password) VALUES (?, ?, ?)", pseudo, email, password)
	if err != nil {
		http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/loghandler", http.StatusSeeOther)
}
func checkCredentials(db *sql.DB, email, password string) (bool, string, error) {
	var id int
	var pseudo string
	var storedPassword string
	row := db.QueryRow("SELECT id, pseudo, password FROM USER WHERE email = ?", email)
	err := row.Scan(&id, &pseudo, &storedPassword)
	if err == sql.ErrNoRows {
		return false, "", nil
	} else if err != nil {
		return false, "", err
	}
	if password == storedPassword {
		return true, pseudo, nil
	}
	return false, "", nil
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
	isAuthorized, username, err := checkCredentials(db, email, password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if isAuthorized {
		http.SetCookie(w, &http.Cookie{
			Name:   "pseudo",
			Value:  username,
			Path:   "/",
			MaxAge: 36000, // Expire après 1 heure
		})
		http.ServeFile(w, r, "templates/selectgame.html")
	} else {
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
	var id int
	var created_by, max_player, name, mode string

	rows, err := db.Query("SELECT id, created_by, max_player, name, mode FROM rooms")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return err
	}
	defer rows.Close()

	fmt.Println("ID | Created By | Max Players | Name | Mode")
	for rows.Next() {
		err = rows.Scan(&id, &created_by, &max_player, &name, &mode)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return err
		}
		fmt.Printf("%d | %s | %s | %s | %s\n", id, created_by, max_player, name, mode)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error with rows: %v", err)
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
func getRooms(db *sql.DB) ([]Room, error) {
	rows, err := db.Query("SELECT id, created_by, max_player, name, mode, id_game FROM ROOMS")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()
	var rooms []Room
	for rows.Next() {
		var room Room
		err = rows.Scan(&room.ID, &room.CreatedBy, &room.MaxPlayer, &room.Name, &room.Mode, &room.GameID)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error with rows: %v", err)
		return nil, err
	}
	return rooms, nil
}
func join(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/room.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	db := database.InitDB()
	defer db.Close()
	rooms, err := getRooms(db)
	if err != nil {
		log.Printf("Error getting rooms: %v", err)
		http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, rooms)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
func createblind(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	r.ParseForm()

	// Get the mode from the form data
	mode := r.Form.Get("mode")

	// Set a cookie with the mode
	http.SetCookie(w, &http.Cookie{
		Name:  "mode",
		Value: mode,
	})

	// Parse the HTML template
	t, err := template.ParseFiles("./templates/create/createblind.html")
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Execute the template
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getRoomByID(db *sql.DB, id int) (Room, error) {
	var room Room
	row := db.QueryRow("SELECT id, created_by, max_player, name, mode FROM ROOMS WHERE ID = ?", id)
	err := row.Scan(&room.ID, &room.CreatedBy, &room.MaxPlayer, &room.Name, &room.Mode)
	if err != nil {
		if err == sql.ErrNoRows {
			return Room{}, fmt.Errorf("room with ID %d not found", id)
		}
		return Room{}, err
	}
	return room, nil
}

func joinRoom(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		fmt.Fprintf(w, "Invalid room ID: %v", err)
		return
	}
	room, err := getRoomByID(db, id)
	if err != nil {
		fmt.Printf("Error getting room details: %v\n", err) // Print the error to the console
		http.Error(w, "Failed to get room details", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("./templates/rooms.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, room)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	db := database.InitDB()
	defer db.Close()

	printRooms(db)
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
	http.HandleFunc("/join-room", func(w http.ResponseWriter, r *http.Request) {
		joinRoom(db, w, r)
	})
	http.HandleFunc("/start-game", func(w http.ResponseWriter, r *http.Request) {
		players := []*blindtest.Player{
			{Name: "Joueur 1"},
			{Name: "Joueur 2"},
		}
		bt := blindtest.NewBlindTest(players)

		// Nombre de tours à jouer
		const roundsToPlay = 3

		// Boucle principale du jeu
		for i := 0; i < roundsToPlay; i++ {
			// Démarrer un nouveau tour de jeu
			err := bt.StartRound()
			if err != nil {
				log.Fatalf("Erreur lors du démarrage du tour de jeu: %v", err)
			}

			// Logique supplémentaire pour gérer le tour de jeu
			// ...
		}
	})
	http.HandleFunc("/addroom", AddRoom)
	http.HandleFunc("/echo", groupieWebsocket.WebsocketHandler)
	http.HandleFunc("/echo2", groupieWebsocket.WebsocketHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("The server is running on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
