package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/Yury132/Golang-Task-1/internal/config"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type Handler struct {
	log zerolog.Logger
}

func New(log zerolog.Logger) *Handler {
	return &Handler{
		log: log,
	}
}

// Вся информация о пользователе из Google
type ViewData struct {
	Title string
}

// Для Google
var (
	googleOauthConfig *oauth2.Config
	// Любая строка
	oauthStateString = "pseudo-random"
	info             ViewData
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := "{\"health\": \"ok\"}"

	response, err := json.Marshal(data)
	if err != nil {
		h.log.Error().Err(err).Msg("filed to marshal response data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// Стартовая страница
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/home_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Авторизация через Google
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	googleOauthConfig = config.SetupConfig()
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google перенаправляет сюда
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		h.log.Error().Err(err).Msg("callback...")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Заполняем структуру инфой из гугла, но не передаем ее на страницу
	info = ViewData{
		Title: string(content),
	}

	tmpl, err := template.ParseFiles("templates/auth_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

// Получаем данные о пользователи из Google
func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}

// Google перенаправляет сюда
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/auth_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, info)

}

// content := { "id": "105118128147454782975", "email": "ivan.ivanov132132@gmail.com", "verified_email": true, "name": "YURIY USYNIN", "given_name": "YURIY", "family_name": "USYNIN", "picture": "https://lh3.googleusercontent.com/a/ACg8ocLJMKT2_vAvctEMY5iygMWj7CzaPLpRvujVH6-hYVJP=s96-c", "locale": "ru" }
