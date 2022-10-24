package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"microblog/storage"
	"microblog/storage/mongo"
	"net/http"
	"strings"
	"testing"
)

var ctx = context.Background()

func TestAPI(t *testing.T) {
	//t.Run("InMemory", func(t *testing.T) {
	//	suite.Run(t, &APISuite{
	//		storage: inmemory.NewStorage(),
	//	})
	//})
	t.Run("Mongo", func(t *testing.T) {
		suite.Run(t, &APISuite{
			storage: mongo.NewStorage("mongodb://localhost:27017", "test_mongo"),
		})
	})
}

type APISuite struct {
	suite.Suite
	storage storage.Storage
	server  *http.Server
	client  http.Client
}

func (s *APISuite) SetupSuite() {
	s.server = NewServer(s.storage, "0.0.0.0:8080")
	go func() {
		log.Printf("Start serving on %s", s.server.Addr)
		log.Fatal(s.server.ListenAndServe())
	}()
}

//func (s *APISuite) TearDownSuite() {
//	s.server.Shutdown(ctx)
//}

func (s *APISuite) TestNotFound() {

	resp, err := s.client.Get("http://localhost:8080/api/v1/posts/funnypostname")

	s.Require().NoError(err)
	s.Require().Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *APISuite) TestCreateAndGet() {
	const postText = "Hola, guapo"
	const authorName = "chico"
	var postId string

	s.Run("CheckCreatePost", func() {
		// when:
		req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/posts", strings.NewReader(fmt.Sprintf(`{"text": "%s"}`, postText)))
		s.Require().NoError(err)
		req.Header.Set("System-Design-User-Id", authorName)
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		rawBody, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var body map[string]string
		s.Require().NoError(json.Unmarshal(rawBody, &body))
		s.Require().NotEmpty(body["id"])
		s.Require().Equal(body["text"], postText)
		s.Require().Equal(body["authorId"], authorName)
		postId = body["id"]
	})

	s.Run("CheckGetPost", func() {
		resp, err := s.client.Get(fmt.Sprintf("http://localhost:8080/api/v1/posts/%s", postId))
		s.Require().NoError(err)
		rawBody, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var body map[string]string
		s.Require().NoError(json.Unmarshal(rawBody, &body))
		s.Require().Equal(body["text"], postText)
		s.Require().Equal(body["authorId"], authorName)
	})
}
