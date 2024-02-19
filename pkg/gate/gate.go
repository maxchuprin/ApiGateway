// запросы к микросервисам системы новостного агрегатора: NewsAgg, Censor, Comments
package gate

import (
	"ApiGateway/pkg/obj"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// делает запрос в сервис новостей и возвращает массив новостей
func GetLatestNews(ctx context.Context, p int) (any, error) {
	//получаем из окружения адрес агрегатора новостей
	newsAggregator := os.Getenv("newsAggregator")
	//запрашиваем агрегатор
	r, err := http.Get(newsAggregator + "/news/" + strconv.Itoa(p*15) + "?requestID=" + getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив новостей+requestID.
	var data = struct {
		Posts     []obj.NewsShortDetailed
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// запрос для добавления нового комментария
func PostComment(ctx context.Context, c obj.Comment) (any, error) {

	//запрос в сервис цензор для выяснения допустимости добавления комментария

	var t = struct{ Text string }{
		Text: c.Text,
	}

	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)

	//получаем из окружения адрес цензора
	cersorService := os.Getenv("cersorService")
	//запрашиваем цензор
	rr, err := http.Post(cersorService+"/check"+"?requestID="+getRequestID(ctx), "application/json", buf)
	if err != nil {
		return nil, err
	}
	defer rr.Body.Close()
	// Проверяем код ответа.
	if !(rr.StatusCode == http.StatusOK) {
		return nil, fmt.Errorf("код ответ сервиса цензурирования при попытке создать комментарий: %d", rr.StatusCode)
	}

	//запрос в сервис комментариев для добавления комментария
	b, err = json.Marshal(c)
	if err != nil {
		return nil, err
	}
	buf = bytes.NewBuffer(b)

	//получаем из окружения адрес сервиса комментариев
	commentsService := os.Getenv("commentsService")

	r, err := http.Post(commentsService+"/add"+"?requestID="+getRequestID(ctx), "application/json", buf)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	// Проверяем код ответа.
	if !(r.StatusCode == http.StatusOK) {
		return nil, fmt.Errorf("код ответ сервиса комментариев при попытке создать новый: %d", r.StatusCode)
	}
	var data = struct {
		RequestID any
	}{}

	// Читаем тело ответа.
	b, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// запрос для получения комментариев к новости
func GetComments(ctx context.Context, id int) ([]obj.Comment, error) {

	//получаем из окружения адрес сервиса комментариев
	commentsService := os.Getenv("commentsService")

	r, err := http.Get(commentsService + "/comments?postID=" + strconv.Itoa(id) + "&requestID=" + getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив комментариев+id запроса
	var data = struct {
		Comments  []obj.Comment
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data.Comments, nil
}

// запрос для получения новости
func GetPost(ctx context.Context, id int) (*obj.NewsFullDetailed, error) {

	//получаем из окружения адрес агрегатора новостей
	newsAggregator := os.Getenv("newsAggregator")
	//запрашиваем агрегатор
	r, err := http.Get(newsAggregator + "/news?postID=" + strconv.Itoa(id) + "&requestID=" + getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON
	var data = struct {
		Post      obj.NewsFullDetailed
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data.Post, nil
}

// отправляет 2 асинхронных запроса - в сервис новостей и сервис комментариев и готовит объект подробной новости
func GetDetailedPost(ctx context.Context, id int) (any, error) {
	c := make(chan interface{}, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		var r commentsResponse
		r.comments, r.err = GetComments(ctx, id)
		c <- r
	}()

	go func() {
		defer wg.Done()
		var r postResponse
		r.post, r.err = GetPost(ctx, id)
		c <- r
	}()

	wg.Wait()
	close(c)

	var r obj.NewsFullDetailed
	var com []obj.Comment

	for m := range c {
		switch m.(type) {
		case commentsResponse:
			a := m.(commentsResponse)
			if a.err != nil {
				return nil, a.err
			}
			com = a.comments
		case postResponse:
			a := m.(postResponse)
			if a.err != nil {
				return nil, a.err
			}
			r = *a.post
		}
	}
	r.Comment = com

	var ans = struct {
		obj.NewsFullDetailed
		RequestID any
	}{
		NewsFullDetailed: r,
		RequestID:        ctx.Value(obj.ContextKey("requestID")),
	}

	return &ans, nil
}

type commentsResponse struct {
	comments []obj.Comment
	err      error
}

type postResponse struct {
	post *obj.NewsFullDetailed
	err  error
}

// запрашивает сервис аггрегатора новостей с поисковым запросом
func SearchPosts(ctx context.Context, searchParam string, pageParam string) (any, error) {

	//получаем из окружения адрес агрегатора новостей
	newsAggregator := os.Getenv("newsAggregator")
	//запрашиваем агрегатор
	reqStr := newsAggregator + "/news?requestID=" + getRequestID(ctx)

	if searchParam != "" {
		reqStr += "&search=" + searchParam
	}

	if pageParam != "" {
		reqStr += "&page=" + pageParam
	}

	r, err := http.Get(reqStr)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Раскодируем JSON в массив новостей + объект пагинации+ id запроса.

	var data = struct {
		Posts      []obj.NewsShortDetailed
		Pagination obj.Pagination
		RequestID  any
	}{}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// получение requestID из контекста
func getRequestID(ctx context.Context) string {
	return strconv.Itoa(ctx.Value(obj.ContextKey("requestID")).(int))
}
