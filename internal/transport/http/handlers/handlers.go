package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/Yury132/Golang-Task-1/internal/models"
	"github.com/rs/zerolog"
)

type Service interface {
	GetUserInfo(state string, code string) ([]byte, error)
	GetUsersList(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log         zerolog.Logger
	oauthConfig *oauth2.Config
	service     Service
}

// Для Google
var (
	// Любая строка
	oauthStateString = "pseudo-random"
	info             models.Content
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
	url := h.oauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google перенаправляет сюда
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// TODO как привести к виду
	content, err := h.service.GetUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		h.log.Error().Err(err).Msg("callback...")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Заполняем структуру инфой из гугла, но не передаем ее на страницу

	//var info models.Content
	if err = json.Unmarshal(content, &info); err != nil {
		h.log.Error().Err(err).Msg("filed to unmarshal struct")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(info.Name)

	tmpl, err := template.ParseFiles("templates/auth_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

// Google перенаправляет сюда
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/auth_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(info.Name)
	tmpl.Execute(w, info)

}

func (h *Handler) GetUsersList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	users, err := h.service.GetUsersList(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error().Err(err).Msg("failed to get users list")
		return
	}

	data, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error().Err(err).Msg("failed to marshal users list")
		return
	}

	w.Write(data)
}

func New(log zerolog.Logger, oauthConfig *oauth2.Config, service Service) *Handler {
	return &Handler{
		log:         log,
		oauthConfig: oauthConfig,
		service:     service,
	}
}

// content := { "id": "105118128147454782975", "email": "ivan.ivanov132132@gmail.com", "verified_email": true, "name": "YURIY USYNIN", "given_name": "YURIY", "family_name": "USYNIN", "picture": "https://lh3.googleusercontent.com/a/ACg8ocLJMKT2_vAvctEMY5iygMWj7CzaPLpRvujVH6-hYVJP=s96-c", "locale": "ru" }
