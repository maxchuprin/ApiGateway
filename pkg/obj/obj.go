// objects of data-model
package obj

type ContextKey string

// тип obj.Post для операций с БД
type NewsFullDetailed struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	Content string // содержание публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
	Comment []Comment
}

type NewsShortDetailed struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
}

type Comment struct {
	ID        int       //ID комментария
	PostID    int       //ID поста (новости) к которой осавлен комментарий
	CommentID int       //ID комментария, к которому привязан данный коммент, если этот комментарий является ответом на другой комментарий
	Text      string    //Текст комментария
	Answers   []Comment //Ответы на данный комментарий
}

//NewsFullDetailed - поля Post сервиса новостей + дерево комментов
//NewsShortDetailed - не все поля Post

type Pagination struct {
	Page         int //текущая страница
	Of           int //всего страниц
	PostsPerPage int //PostsPerPage
}
