package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"microblog/storage"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_PAGE_SIZE = 10
const MINIMUM_PAGE_SIZE = 1
const MAXMIMUM_PAGE_SIZE = 100

type HTTPHandler struct {
	storage storage.Storage
}

type CreatePostRequest struct {
	Text string `json:"text"`
}

func HandleRoot(rw http.ResponseWriter, r *http.Request) {
	_, err := rw.Write([]byte("Hola, guapo"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleCreatePost(rw http.ResponseWriter, r *http.Request) {
	var requestData CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	userId := r.Header.Get("System-Design-User-Id")
	re := regexp.MustCompile(`[A-Za-z0-9_\-]+`)
	if !re.MatchString(userId) {
		http.Error(rw, "Токен пользователя отсутствует в запросе, или передан в неверном формате.", http.StatusUnauthorized)
		return
	}

	post := storage.Post{
		Id:        uuid.New().String(),
		Text:      requestData.Text,
		AuthorId:  userId,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err = h.storage.AddPost(r.Context(), post)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(post)

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(rawResponse)
}

func (h *HTTPHandler) HandleGetPost(rw http.ResponseWriter, r *http.Request) {
	postId := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	post, err := h.storage.GetPost(r.Context(), postId)

	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusNotFound)
		return
	}

	rawResponse, _ := json.Marshal(post)

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(rawResponse)
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	userId := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	userId = strings.TrimSuffix(userId, "/posts")

	page := r.URL.Query().Get("page")

	size := DEFAULT_PAGE_SIZE
	querySize := r.URL.Query().Get("size")
	if querySize != "" {
		var err error
		size, err = strconv.Atoi(querySize)
		if err != nil || size < MINIMUM_PAGE_SIZE || size > MAXMIMUM_PAGE_SIZE {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	posts, err := h.storage.GetPostsByUser(r.Context(), userId, page, size)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(posts)

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(rawResponse)
}

func NewServer(storage storage.Storage) *http.Server {
	r := mux.NewRouter()

	handler := &HTTPHandler{
		storage: storage,
	}

	r.HandleFunc("/", HandleRoot).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId}", handler.HandleGetPost).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{userId}/posts", handler.HandleGetUserPosts).Methods(http.MethodGet)

	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return server
}
