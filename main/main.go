package main

import (
    "fmt"
    "html/template"
    "log"
    "database/sql"
    "errors"
    "net/http"
	_ "github.com/mattn/go-sqlite3"

    database "groupie-tracker/bdd"
)

// import (
//     "../controller"
// )

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

    email := r.FormValue("email")
    password := r.FormValue("password")

    db := database.InitDB()
    defer db.Close()

    // Vérifiez si l'utilisateur existe déjà
    var id int
    row := db.QueryRow("SELECT id FROM USER WHERE email = ?", email)
    err := row.Scan(&id)
    if err != sql.ErrNoRows {
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "User already exists.")
        return
    }

    // Insérez l'utilisateur dans la base de données
    _, err = db.Exec("INSERT INTO USER (email, password) VALUES (?, ?)", email, password)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Affichez les informations de l'utilisateur dans la console
    fmt.Println("Email:", email, "Password:", password)

    fmt.Fprintf(w, "Registration successful. Please log in again.")
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
        http.Redirect(w, r, "/test", http.StatusSeeOther)
    } else {
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


    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    fmt.Println("The server is running on port :8081")
    err := http.ListenAndServe(":8081", nil)
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    //controller.TestDatabase()
}