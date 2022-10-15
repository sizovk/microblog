package httpapi

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"sync"
	"time"
)

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

type HTTPHandler struct {
	mu sync.RWMutex
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
	_, err := rw.Write([]byte("CreatePost"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleGetPost(rw http.ResponseWriter, r *http.Request) {
	postId := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	_, err := rw.Write([]byte("GetPosts " + postId))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	userId := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	userId = strings.TrimSuffix(userId, "/posts")
	_, err := rw.Write([]byte("GetUserPosts " + userId))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func NewServer() *http.Server {
	r := mux.NewRouter()

	handler := NewHTTPHandler()

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
