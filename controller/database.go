package controller

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
	//il faut obtenir sqlite avec: go get -u github.com/mattn/go-sqlite3
)

func TestDatabase() {
    // Connexion à la base de données
    db, err := sql.Open("sqlite3", "./bdd/groupie_tracker.db")
    if err != nil {
        fmt.Println("Erreur lors de la connexion à la base de données:", err)
        return
    }
    defer db.Close()

    // Tester la connexion
    err = db.Ping()
    if err != nil {
        fmt.Println("Erreur lors du test de connexion à la base de données:", err)
        return
    }
    fmt.Println("Connexion à la base de données réussie")

    // Exécuter une requête de test
    rows, err := db.Query("SELECT * FROM USER")
    if err != nil {
        fmt.Println("Erreur lors de l'exécution de la requête:", err)
        return
    }
    defer rows.Close()

    // Parcourir les résultats de la requête
    fmt.Println("Utilisateurs:")
    for rows.Next() {
        var id int
        var pseudo, email, password string
        err = rows.Scan(&id, &pseudo, &email, &password)
        if err != nil {
            fmt.Println("Erreur lors de la lecture des données:", err)
            return
        }
        fmt.Printf("ID: %d, Pseudo: %s, Email: %s, Mot de passe: %s\n", id, pseudo, email, password)
    }
    if err = rows.Err(); err != nil {
        fmt.Println("Erreur lors de l'itération sur les résultats:", err)
        return
    }
}
