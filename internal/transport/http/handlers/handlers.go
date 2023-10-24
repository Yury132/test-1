package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/Yury132/Golang-Task-1/internal/models"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog"
)

type Service interface {
	GetUserInfo(state string, code string) ([]byte, error)
	GetUsersList(ctx context.Context) ([]models.User, error)
	// Проверка на существование пользователя
	CheckUser(ctx context.Context, email string) (bool, error)
	// Создание нового пользователя
	CreateUser(ctx context.Context, name string, email string) error
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
	store            = sessions.NewCookieStore([]byte("super-secret-key"))
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

// Google перенаправляет сюда, когда пользователь успешно авторизовался, создаем сессию
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// Получаем данные из гугла
	content, err := h.service.GetUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		h.log.Error().Err(err).Msg("callback...")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Заполняем info, но не передаем ее на страницу
	if err = json.Unmarshal(content, &info); err != nil {
		h.log.Error().Err(err).Msg("filed to unmarshal struct")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Обращаемся к БД и узнаем есть ли такой пользователь в системе
	alive, err := CheckUser(r, h, info.Email)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to check user...")
		alive = false
	}
	// Если пользователя нет в БД, создаем его
	if !alive {
		err = CreateUser(r, h, info.Name, info.Email)
		if err != nil {
			h.log.Error().Err(err).Msg("failed to create user...")
		}
	}

	// Задаем жизнь сессии в секундах
	// 10 мин
	store.Options = &sessions.Options{
		MaxAge: 60 * 10,
	}
	// Создаем сессию
	session, err := store.Get(r, "session-name")
	if err != nil {
		h.log.Error().Err(err).Msg("session create failed")
	}
	// Устанавливаем значения в сессию
	// Сохраняем данные пользователя
	session.Values["authenticated"] = true
	session.Values["Name"] = info.Name
	session.Values["Email"] = info.Email
	session.Save(r, w)

	tmpl, err := template.ParseFiles("templates/auth_page.html")
	if err != nil {
		h.log.Error().Err(err).Msg("filed to show home page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Информация о пользователе
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {

	// Получаем сессию
	session, err := store.Get(r, "session-name")
	if err != nil {
		h.log.Error().Err(err).Msg("session failed")
		//w.WriteHeader(http.StatusInternalServerError)
		//return
	}

	// Проверяем, что пользователь залогинен
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		// Если нет
		// w.WriteHeader(http.StatusUnauthorized)
		// w.Header().Set("Content-Type", "application/json")
		// resp := make(map[string]string)
		// resp["сообщение"] = "Вы не авторизованы..."
		// jsonResp, err := json.Marshal(resp)
		// if err != nil {
		// 	h.log.Error().Err(err).Msg("Error happened in JSON marshal")
		// }
		// w.Write(jsonResp)
		//
		tmpl, err := template.ParseFiles("templates/error.html")
		w.WriteHeader(http.StatusUnauthorized)
		if err != nil {
			h.log.Error().Err(err).Msg("filed to show error page")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else {
		// Если да
		// Читаем данные из сессии
		info.Name = session.Values["Name"].(string)
		info.Email = session.Values["Email"].(string)

		tmpl, err := template.ParseFiles("templates/auth_page.html")
		if err != nil {
			h.log.Error().Err(err).Msg("filed to show home page")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, info)
	}
}

// Выход из системы, удаление сессии
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		h.log.Error().Err(err).Msg("session failed")
	}
	// Удаляем сессию
	session.Options.MaxAge = -1
	session.Save(r, w)
	// Переадресуем пользователя на страницу логина
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

// Проверка на существование пользователя
func CheckUser(r *http.Request, h *Handler, email string) (bool, error) {
	check, err := h.service.CheckUser(r.Context(), email)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to check user")
		return false, err
	}
	return check, nil
}

// Создание нового пользователя
func CreateUser(r *http.Request, h *Handler, name string, email string) error {
	err := h.service.CreateUser(r.Context(), name, email)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to create user")
		return err
	}
	return nil
}

func New(log zerolog.Logger, oauthConfig *oauth2.Config, service Service) *Handler {
	return &Handler{
		log:         log,
		oauthConfig: oauthConfig,
		service:     service,
	}
}
