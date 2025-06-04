package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var baseURL = "http://localhost:8080"
var postURL = "http://localhost:8081/posts"
var statsURL = "http://localhost:8083/stats"

func generateRandomUser() (login, password, email string) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(100000)
	login = fmt.Sprintf("user%d", n)
	password = "password"
	email = fmt.Sprintf("user%d@example.com", n)
	return
}

func mustRequest(method, url, body string) *http.Request {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestEndToEnd_Register_Post_Like_Stats(t *testing.T) {
	client := &http.Client{}
	var cookies []*http.Cookie

	login, password, email := generateRandomUser()

	registerBody := fmt.Sprintf(`{"login":"%s","password":"%s","email":"%s"}`, login, password, email)
	req, _ := http.NewRequest("POST", baseURL+"/register", strings.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := client.Do(req)

	authBody := fmt.Sprintf(`{"login":"%s","password":"%s"}`, login, password)
	req, _ = http.NewRequest("POST", baseURL+"/auth", strings.NewReader(authBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	cookies = resp.Cookies()

	postReqBody := `{"title": "E2E Post", "description": "Hello", "creatorId": 1, "isPrivate": false, "tags": ["e2e"]}`
	req, _ = http.NewRequest("POST", postURL, strings.NewReader(postReqBody))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, _ = client.Do(req)
	bodyBytes, _ := io.ReadAll(resp.Body)
	var createdPost map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &createdPost)
	postID := int(createdPost["id"].(float64))

	likeURL := baseURL + "/posts/" + strconv.Itoa(postID) + "/like"
	req, _ = http.NewRequest("POST", likeURL, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client.Do(req)

	time.Sleep(2 * time.Second)

	req, _ = http.NewRequest("GET", statsURL+"/top/posts?by=likes", nil)
	resp, _ = client.Do(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), strconv.Itoa(postID))
}

func TestEndToEnd_PrivatePostPermission(t *testing.T) {
	client := &http.Client{}
	var cookies []*http.Cookie

	login1, password1, email1 := generateRandomUser()
	registerBody := fmt.Sprintf(`{"login":"%s","password":"%s","email":"%s"}`, login1, password1, email1)
	authBody := fmt.Sprintf(`{"login":"%s","password":"%s"}`, login1, password1)
	client.Do(mustRequest("POST", baseURL+"/register", registerBody))
	resp, _ := client.Do(mustRequest("POST", baseURL+"/auth", authBody))
	cookies = resp.Cookies()

	postReqBody := `{"title": "Private Post", "description": "hidden", "creatorId": 1, "isPrivate": true, "tags": []}`
	req, _ := http.NewRequest("POST", postURL, strings.NewReader(postReqBody))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, _ = client.Do(req)
	bodyBytes, _ := io.ReadAll(resp.Body)
	var created map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &created)
	postID := int(created["id"].(float64))

	login2, password2, email2 := generateRandomUser()
	registerBody2 := fmt.Sprintf(`{"login":"%s","password":"%s","email":"%s"}`, login2, password2, email2)
	authBody2 := fmt.Sprintf(`{"login":"%s","password":"%s"}`, login2, password2)
	client.Do(mustRequest("POST", baseURL+"/register", registerBody2))
	resp, _ = client.Do(mustRequest("POST", baseURL+"/auth", authBody2))
	otherCookies := resp.Cookies()

	req, _ = http.NewRequest("GET", postURL+"/"+strconv.Itoa(postID)+"?requesterId=2", nil)
	for _, c := range otherCookies {
		req.AddCookie(c)
	}
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestEndToEnd_CommentAndTopUsers(t *testing.T) {
	client := &http.Client{}
	var cookies []*http.Cookie

	login, password, email := generateRandomUser()
	registerBody := fmt.Sprintf(`{"login":"%s","password":"%s","email":"%s"}`, login, password, email)
	authBody := fmt.Sprintf(`{"login":"%s","password":"%s"}`, login, password)
	client.Do(mustRequest("POST", baseURL+"/register", registerBody))
	resp, _ := client.Do(mustRequest("POST", baseURL+"/auth", authBody))
	cookies = resp.Cookies()

	postReqBody := `{"title": "Commentable", "description": "pls comment", "creatorId": 1, "isPrivate": false, "tags": []}`
	req, _ := http.NewRequest("POST", postURL, strings.NewReader(postReqBody))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, _ = client.Do(req)
	bodyBytes, _ := io.ReadAll(resp.Body)
	var created map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &created)
	postID := int(created["id"].(float64))

	commentBody := `{"clientId":1, "content":"great post!"}`
	req, _ = http.NewRequest("POST", baseURL+"/posts/"+strconv.Itoa(postID)+"/comments", strings.NewReader(commentBody))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client.Do(req)

	time.Sleep(2 * time.Second)

	resp, _ = client.Get(statsURL + "/top/users?by=comments")
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"user_id":1`)

}
