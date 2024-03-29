package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func HandlePing(rw http.ResponseWriter, r *http.Request) {
	_, err := rw.Write([]byte("Pong"))
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
		Id:             primitive.NewObjectID().Hex(),
		Text:           requestData.Text,
		AuthorId:       userId,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		LastModifiedAt: time.Now().UTC().Format(time.RFC3339),
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
			http.Error(rw, "Некорректный запрос, например, из-за некорректного токена страницы.", http.StatusBadRequest)
			return
		}
	}

	posts, err := h.storage.GetPostsByUser(r.Context(), userId, page, size)

	if err != nil {
		http.Error(rw, "Некорректный запрос, например, из-за некорректного токена страницы.", http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(posts)

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(rawResponse)
}

func (h *HTTPHandler) HandlePatchPost(rw http.ResponseWriter, r *http.Request) {
	var requestData CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	userId := r.Header.Get("System-Design-User-Id")
	re := regexp.MustCompile(`[A-Za-z0-9_\-]+`)
	if !re.MatchString(userId) {
		http.Error(rw, "Пользователь не аутентифирован", http.StatusUnauthorized)
		return
	}

	postId := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	post, err := h.storage.GetPost(r.Context(), postId)

	if err != nil {
		http.Error(rw, "Поста с указанным идентификатором не существует", http.StatusNotFound)
		return
	}

	if post.AuthorId != userId {
		http.Error(rw, "Пост не может быть отредактирован, т.к. опубликован другим пользователем.", http.StatusForbidden)
		return
	}

	post.Text = requestData.Text
	post.LastModifiedAt = time.Now().UTC().Format(time.RFC3339)

	err = h.storage.PatchPost(r.Context(), post)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(post)

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(rawResponse)
}

func NewServer(storage storage.Storage, address string) *http.Server {
	r := mux.NewRouter()

	handler := &HTTPHandler{
		storage: storage,
	}

	r.HandleFunc("/", HandleRoot).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/maintenance/ping", HandlePing).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId}", handler.HandleGetPost).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{userId}/posts", handler.HandleGetUserPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts/{postId}", handler.HandlePatchPost).Methods(http.MethodPatch)

	server := &http.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return server
}
