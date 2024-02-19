// api сервера
package api

import (
	"ApiGateway/pkg/gate"
	"ApiGateway/pkg/obj"
	"encoding/json"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API сервиса.
type API struct {
	r *mux.Router // маршрутизатор запросов
}

// Конструктор API.
func New() *API {
	api := API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Router возвращает маршрутизатор запросов.
func (api *API) Router() *mux.Router {
	return api.r
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	// получить n последних новостей
	api.r.HandleFunc("/news/latest", api.posts).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/search", api.search).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/post", api.postByID).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/comment", api.addComment).Methods(http.MethodPost, http.MethodOptions)
	//заголовок ответа
	api.r.Use(api.HeadersMiddleware)
	api.r.Use(api.RequestIDMiddleware)
	api.r.Use(api.LoggingMiddleware)
}

// posts возвращает n-ую страницу (страница = 10 постов) новейших новостей в зависимости от параметра page=n
func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	// Считывание параметра page строки запроса.

	// если параметр был передан, вернется строка со значением.
	// Если не был - в переменной будет пустая строка
	pageParam := r.URL.Query().Get("page")
	// параметр page - это число, поэтому нужно сконвертировать
	// строку в число при помощи пакета strconv
	var page int
	var err error

	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		page = 1
	}
	// Получение данных из сервиса новостей
	o, err := gate.GetLatestNews(r.Context(), page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(o)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

var FullNews = []obj.NewsFullDetailed{
	{ID: 1,
		Title:   "Конец света 2089",
		Content: "Очередное описание фейковой новости",
		PubTime: 1696349769,
		Link:    "www.rbc.ru"},
	{ID: 2,
		Title:   "Землятресение в Алматы",
		Content: "В Алматы произошло сильное землятресение в 5 баллов",
		PubTime: 1696349770,
		Link:    "www.kommersant.ru",
		Comment: []obj.Comment{}},
}

var ShortNews = []obj.NewsShortDetailed{
	{ID: 1,
		Title:   "Важная новость",
		PubTime: 1696349769,
		Link:    "www.rbc.ru"},
	{ID: 2,
		Title:   "Ещё какая то важная новость",
		PubTime: 1696349770,
		Link:    "www.kommersant.ru"},
}

// Фильтр или поиск новостей.
// Для данного метода параметры:
// contains = слово -  совпадение слова в заголовке новости
// dateafter, datebefore = UNIX time -  выбор диапазона дат,
// notcontains = слово -  слова в заголовке, которые исключить,
// sort = date/name - выбор поля для сортировки (по дате, по названию).
func (api *API) search(w http.ResponseWriter, r *http.Request) {
	// Считывание параметров фильтра page строки запроса.
	searchParam := r.URL.Query().Get("search")
	pageParam := r.URL.Query().Get("page")

	// Получение данных из сервиса новостей
	ans, err := gate.SearchPosts(r.Context(), searchParam, pageParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(ans)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

// возвращает детальную новость по postID
func (api *API) postByID(w http.ResponseWriter, r *http.Request) {
	// Считывание параметра  строки запроса.
	idParam := r.URL.Query().Get("postID")

	id, err := strconv.Atoi(idParam)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получение данных из сервиса новостей и сервиса комментариев

	n, err := gate.GetDetailedPost(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(n)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

// метод добавления комментария
// Принимает ID новости postID или ID родительского комментария commentID и текст комментария в теле запроса
func (api *API) addComment(w http.ResponseWriter, r *http.Request) {

	var c obj.Comment
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//тут отправка запроса на создание комментария  в сервис комментариев
	data, err := gate.PostComment(r.Context(), c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(data)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}
