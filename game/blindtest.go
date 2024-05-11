package blindtest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

// SpotifyClient représente un client pour interagir avec l'API Spotify
type SpotifyClient struct {
	ClientID     string
	ClientSecret string
}

// Initialisez vos identifiants d'API Spotify
const (
	ClientID     = "0d37976efb0445168156a2f992f84af6"
	ClientSecret = "5d2b98f826824e39a7e25c2acdd6eb9f"
)

// Track représente une piste musicale
type Track struct {
	Name     string `json:"name"`
	AudioURL string `json:"audio_url"`
}

// BlindTest représente le jeu de blindtest
type BlindTest struct {
	Players      []*Player
	CurrentRound *Round
	Playlist     *Playlist
	Spotify      *SpotifyClient
}

// Playlist représente une liste de chansons pour le jeu de blindtest
type Playlist struct {
	Songs []*Song // Changement ici
}

// Song représente une chanson dans la playlist
type Song struct {
	Title    string
	Artist   string
	AudioURL string
}

// Player représente un joueur dans le jeu de blindtest
type Player struct {
	Name  string
	Score int
}

// Round représente un tour de jeu
type Round struct {
	Song *Song
}

// NewBlindTest crée une nouvelle instance du jeu de blindtest
func NewBlindTest(players []*Player) *BlindTest {
	spotify := &SpotifyClient{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	return &BlindTest{
		Players:  players,
		Spotify:  spotify,
		Playlist: &Playlist{},
	}
}

// StartRound démarre un nouveau tour de jeu
func (bt *BlindTest) StartRound() error {
	// Sélectionner une chanson aléatoire de la playlist
	song, err := bt.getRandomSongFromPlaylist()
	if err != nil {
		return fmt.Errorf("erreur lors de la sélection de la chanson aléatoire: %v", err)
	}
	// Créer le tour avec la chanson sélectionnée
	bt.CurrentRound = &Round{
		Song: song,
	}
	return nil
}

// // loadPlaylistFromSpotify charge la playlist depuis Spotify
//
//	func (bt *BlindTest) loadPlaylistFromSpotify() error {
//		// Générer le jeton d'accès Spotify
//		accessToken, err := bt.generateAccessToken()
//		if err != nil {
//			return fmt.Errorf("erreur lors de la génération du jeton d'accès Spotify: %v", err)
//		}
//		// Récupérer toutes les chansons de la playlist Spotify
//		err = bt.getAllSongsFromSpotifyPlaylist(accessToken, "37i9dQZF1EIZpmEBIVDg9r")
//		if err != nil {
//			return fmt.Errorf("erreur lors du chargement des chansons de la playlist Spotify: %v", err)
//		}
//		return nil
//	}
//
// getRandomSongFromPlaylist récupère une chanson aléatoire de la playlist en utilisant l'API Spotify
func (bt *BlindTest) getRandomSongFromPlaylist() (*Song, error) {
	// Générer le jeton d'accès Spotify
	accessToken, err := bt.generateAccessToken()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération du jeton d'accès Spotify: %v", err)
	}
	// Utiliser le jeton d'accès pour récupérer une chanson aléatoire de la playlist
	randomSong, err := bt.getSongFromSpotifyPlaylist(accessToken)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de la chanson aléatoire de la playlist Spotify: %v", err)
	}
	return randomSong, nil
}

// generateAccessToken génère le jeton d'accès Spotify
func (bt *BlindTest) generateAccessToken() (string, error) {
	// Encodez le client ID et le client secret pour obtenir le header Authorization
	authHeader := base64.StdEncoding.EncodeToString([]byte(bt.Spotify.ClientID + ":" + bt.Spotify.ClientSecret))
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

// getSongFromSpotifyPlaylist récupère une chanson aléatoire de la playlist Spotify
func (bt *BlindTest) getSongFromSpotifyPlaylist(accessToken string) (*Song, error) {
	playlistID := "37i9dQZF1EIZpmEBIVDg9r" // ID de votre playlist
	// Récupérer la liste de toutes les chansons dans la playlist
	allSongs, err := bt.getAllSongsFromSpotifyPlaylist(accessToken, playlistID)
	if err != nil {
		return nil, err
	}
	// Sélectionner une chanson aléatoire parmi la liste
	randomIndex := rand.Intn(len(allSongs))
	return &allSongs[randomIndex], nil
}

// getAllSongsFromSpotifyPlaylist récupère toutes les chansons de la playlist Spotify
func (bt *BlindTest) getAllSongsFromSpotifyPlaylist(accessToken, playlistID string) ([]Song, error) {
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
		return nil, fmt.Errorf("statut de réponse non valide : %s", resp.Status)
	}
	// Lire le corps de la réponse JSON
	var data struct {
		Items []struct {
			Track struct {
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"track"`
		} `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	// Construire la liste des chansons à partir des données de réponse
	var songs []Song
	for _, item := range data.Items {
		artists := make([]string, len(item.Track.Artists))
		for i, artist := range item.Track.Artists {
			artists[i] = artist.Name
		}
		songs = append(songs, Song{
			Title:    item.Track.Name,
			Artist:   strings.Join(artists, ", "),
			AudioURL: item.Track.ExternalURLs.Spotify,
		})
	}
	return songs, nil
}

// Ajoutez cette fonction à votre fichier blindtest.go
func (bt *BlindTest) RandomSongHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode de la requête est GET
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	// Sélectionner une chanson aléatoire
	err := bt.StartRound()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de la sélection de la chanson aléatoire: %v", err), http.StatusInternalServerError)
		return
	}
	// Renvoyer les informations sur la chanson aléatoire sous forme de réponse JSON
	response := struct {
		Title    string `json:"title"`
		Artist   string `json:"artist"`
		AudioURL string `json:"audio_url"`
	}{
		Title:    bt.CurrentRound.Song.Title,
		Artist:   bt.CurrentRound.Song.Artist,
		AudioURL: bt.CurrentRound.Song.AudioURL,
	}
	// Encoder la réponse en JSON et l'écrire dans le corps de la réponse HTTP
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de l'encodage JSON de la réponse: %v", err), http.StatusInternalServerError)
		return
	}
}
