package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ClientID     = "0d37976efb0445168156a2f992f84af6"
	ClientSecret = "5d2b98f826824e39a7e25c2acdd6eb9f"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Générer le jeton d'accès Spotify
	accessToken, err := generateAccessToken()
	if err != nil {
		fmt.Println("Erreur lors de la génération du jeton d'accès:", err)
		return
	}

	// Utiliser le jeton d'accès pour récupérer une chanson aléatoire de la playlist
	randomSong, err := getRandomSongFromPlaylist(accessToken)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de la chanson aléatoire de la playlist:", err)
		return
	}

	fmt.Println("Chanson aléatoire de la playlist récupérée :", randomSong)
}

func generateAccessToken() (string, error) {
	// Encodez le client ID et le client secret pour obtenir le header Authorization
	authHeader := base64.StdEncoding.EncodeToString([]byte(ClientID + ":" + ClientSecret))

	// Préparez les données de la requête POST
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	// Créez la requête HTTP POST
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Envoyez la requête et traitez la réponse
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func getRandomSongFromPlaylist(accessToken string) (string, error) {
	playlistID := "4OZ02mQrmS1LU8bkG09vq7" // ID de votre playlist

	// Récupérer la liste de toutes les chansons dans la playlist
	allSongs, err := getAllSongsFromPlaylist(accessToken, playlistID)
	if err != nil {
		return "", err
	}

	// Sélectionner une chanson aléatoire parmi la liste
	randomIndex := rand.Intn(len(allSongs))
	return allSongs[randomIndex], nil
}

func getAllSongsFromPlaylist(accessToken, playlistID string) ([]string, error) {
	// Créer la requête GET pour récupérer les chansons de la playlist
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID), nil)
	if err != nil {
		return nil, err
	}

	// Ajouter l'en-tête d'autorisation avec le jeton d'accès
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Envoyer la requête et récupérer la réponse
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Vérifier le code de statut de la réponse
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Statut de réponse non valide : %s", resp.Status)
	}

	// Lire le corps de la réponse JSON
	var data struct {
		Items []struct {
			Track struct {
				Name string `json:"name"`
			} `json:"track"`
		} `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	// Extraire le nom de chaque chanson de la playlist
	var songs []string
	for _, item := range data.Items {
		songs = append(songs, item.Track.Name)
	}

	return songs, nil
}
