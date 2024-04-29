package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	ClientID           = "0d37976efb0445168156a2f992f84af6"
	ClientSecret       = "5d2b98f826824e39a7e25c2acdd6eb9f"
	ClientIDGenius     = "10rg6i4uaUtqnTueftf6fMA-NR2GM89eJ2FyFWv-h1Z-7-irfka9rtscEkxfD3mI"
	ClientSecretGenius = "Ohw2qk9Q3ydzxigftO_HQ5pktONaw2NdJplYTW6bsAbTsMRevuYe3h1VHASaUENidVzRoqoDkmhD43hkz55G4Q"
	ClientIDAccess     = "0usQQiAb3EUV0A3cogy-FzgcgX5NcU04lw8gFGgL-AkJd_VemfO9e9fMYjrzlvFA"
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

	// Récupérer les paroles de la chanson aléatoire
	err = getLyrics(randomSong, ClientIDAccess)
	if err != nil {
		fmt.Println("Erreur lors de la récupération des paroles de la chanson:", err)
		return
	}
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
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"track"`
		} `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	// Extraire le nom de chaque chanson et l'artiste de la playlist
	var songs []string
	for _, item := range data.Items {
		if len(item.Track.Artists) > 0 {
			songs = append(songs, fmt.Sprintf("Song: %s, Artist: %s", item.Track.Name, item.Track.Artists[0].Name))
		} else {
			songs = append(songs, fmt.Sprintf("Song: %s, Artist: Unknown", item.Track.Name))
		}
	}

	return songs, nil
}

func checkLyricsAvailability(song, artist string, accessToken string) (bool, error) {
	// Créer une requête GET à l'API Genius pour rechercher les paroles de la chanson
	apiUrl := fmt.Sprintf("https://api.genius.com/search?q=%s %s", url.QueryEscape(song), url.QueryEscape(artist))

	// Créer un client HTTP
	client := &http.Client{}

	// Créer une requête GET
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return false, err
	}

	// Ajouter l'en-tête d'autorisation avec le jeton d'accès pour Genius
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Envoyer la requête et obtenir une réponse
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Décoder la réponse JSON
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, err
	}

	// Extraire l'objet "response" du JSON
	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return false, errors.New("unexpected response format from Genius API")
	}

	// Extraire le tableau "hits" de la réponse
	hits, ok := response["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return false, nil // Les paroles ne sont pas disponibles
	}

	// Les paroles sont disponibles
	return true, nil
}

func getLyrics(song string, accessToken string) error {
	// Créer une requête GET à l'API Genius pour rechercher les paroles de la chanson
	apiURL := fmt.Sprintf("https://api.genius.com/search?q=%s", url.QueryEscape(song))

	// Créer un client HTTP
	client := &http.Client{}

	// Créer une requête GET
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}

	// Ajouter l'en-tête d'autorisation avec le jeton d'accès pour Genius
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Envoyer la requête et obtenir une réponse
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Décoder la réponse JSON
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	// Extraire l'objet "response" du JSON
	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return errors.New("unexpected response format from Genius API")
	}

	// Extraire le tableau "hits" de la réponse
	hits, ok := response["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return errors.New("no lyrics found for the song")
	}

	// Extraire le premier hit
	hit := hits[0].(map[string]interface{})

	// Extraire l'objet "result" du hit
	result, ok := hit["result"].(map[string]interface{})
	if !ok {
		return errors.New("unexpected response format from Genius API")
	}

	// Extraire l'URL de la page des paroles
	lyricsURL, ok := result["url"].(string)
	if !ok {
		return errors.New("unexpected response format from Genius API")
	}

	// Effectuer une requête GET pour récupérer la page des paroles
	resp, err = client.Get(lyricsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Analyser la page des paroles
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Extraire les paroles de la page
	lyrics := doc.Find(".lyrics").Text()

	// Imprimer les paroles
	fmt.Println("Paroles de la chanson:", lyrics)

	return nil
}
