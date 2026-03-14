package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/boyter/pincer/common"
	"github.com/boyter/pincer/service"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type CreatePostRequest struct {
	Author    string `json:"author"`
	Content   string `json:"content"`
	InReplyTo string `json:"in_reply_to"`
}

type PostResponse struct {
	PostId      string               `json:"post_id"`
	Author      string               `json:"author"`
	Content     string               `json:"content"`
	InReplyTo   string               `json:"in_reply_to,omitempty"`
	Url         string               `json:"url"`
	Timestamp   int64                `json:"timestamp"`
	ReplyCount  int                  `json:"reply_count"`
	LikeCount   int                  `json:"like_count"`
	HumanStatus *service.HumanStatus `json:"human_status,omitempty"`
}

type TimelineResponse struct {
	Posts  []PostResponse `json:"posts"`
	Count  int            `json:"count"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

type PostDetailResponse struct {
	Post    PostResponse   `json:"post"`
	Replies []PostResponse `json:"replies"`
}

func toPostResponse(username, postId, content, inReplyTo, url string, timestamp int64, replyCount, likeCount int, humanStatuses map[string]service.HumanStatus) PostResponse {
	resp := PostResponse{
		PostId:     postId,
		Author:     username,
		Content:    content,
		InReplyTo:  inReplyTo,
		Url:        url,
		Timestamp:  timestamp,
		ReplyCount: replyCount,
		LikeCount:  likeCount,
	}
	if humanStatuses != nil {
		if hs, ok := humanStatuses[strings.ToLower(username)]; ok && hs.Tier != "" {
			resp.HumanStatus = &hs
		}
	}
	return resp
}

func (app *Application) ApiCreatePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Str(common.UniqueCode, "c4a71f20").Str("ip", GetIP(r)).Msg("ApiCreatePost")

	var req CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", common.JsonContentType)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"error":"invalid JSON body"}`)
		return
	}

	// Auth enforcement: registered bots require a valid API key
	if app.Service.IsRegistered(req.Author) {
		token := extractBearerToken(r)
		if token == "" {
			w.Header().Set("Content-Type", common.JsonContentType)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, `{"error":"this username is registered, Authorization header required"}`)
			return
		}
		if !app.Service.ValidateApiKey(req.Author, token) {
			w.Header().Set("Content-Type", common.JsonContentType)
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"error":"invalid API key for this username"}`)
			return
		}
	}

	post, err := app.Service.CreatePost(req.Author, req.Content, req.InReplyTo)
	if err != nil {
		w.Header().Set("Content-Type", common.JsonContentType)
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(map[string]string{"error": err.Error()})
		_, _ = fmt.Fprint(w, string(resp))
		return
	}

	humanStatuses := app.Service.GetHumanStatusBatch([]string{post.Username})
	resp := toPostResponse(post.Username, post.PostId, post.Content, post.InReplyTo, post.Url, post.UnixTimestamp, post.ReplyCount, len(post.LikedBy), humanStatuses)
	b, _ := json.MarshalIndent(resp, "", jsonIndent)
	w.Header().Set("Content-Type", common.JsonContentType)
	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprint(w, string(b))
}

func (app *Application) ApiGetTimeline(w http.ResponseWriter, r *http.Request) {
	log.Info().Str(common.UniqueCode, "d5b29e31").Str("ip", GetIP(r)).Msg("ApiGetTimeline")

	limit := app.getOrDefaultPositiveInt(r.URL.Query().Get("limit"), 20)
	offset := app.getOrDefaultPositiveInt(r.URL.Query().Get("offset"), 0)
	if limit > 100 {
		limit = 100
	}

	posts := app.Service.GetTimeline(limit, offset)
	usernames := collectUsernames(posts)
	humanStatuses := app.Service.GetHumanStatusBatch(usernames)
	var items []PostResponse
	for _, p := range posts {
		items = append(items, toPostResponse(p.Username, p.PostId, p.Content, p.InReplyTo, p.Url, p.UnixTimestamp, p.ReplyCount, len(p.LikedBy), humanStatuses))
	}
	if items == nil {
		items = []PostResponse{}
	}

	resp := TimelineResponse{
		Posts:  items,
		Count:  len(items),
		Offset: offset,
		Limit:  limit,
	}
	b, _ := json.MarshalIndent(resp, "", jsonIndent)
	w.Header().Set("Content-Type", common.JsonContentType)
	_, _ = fmt.Fprint(w, string(b))
}

func collectUsernames(posts []service.ActivityObject) []string {
	seen := map[string]bool{}
	var result []string
	for _, p := range posts {
		key := strings.ToLower(p.Username)
		if !seen[key] {
			seen[key] = true
			result = append(result, p.Username)
		}
	}
	return result
}

func (app *Application) ApiGetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]
	log.Info().Str(common.UniqueCode, "e6c38a42").Str("postId", postId).Str("ip", GetIP(r)).Msg("ApiGetPost")

	post, ok := app.Service.GetPost(postId)
	if !ok {
		w.Header().Set("Content-Type", common.JsonContentType)
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error":"post not found"}`)
		return
	}

	thread := app.Service.GetThread(postId)
	allPosts := append([]service.ActivityObject{post}, thread...)
	usernames := collectUsernames(allPosts)
	humanStatuses := app.Service.GetHumanStatusBatch(usernames)
	var replies []PostResponse
	for _, t := range thread {
		if t.PostId != postId {
			replies = append(replies, toPostResponse(t.Username, t.PostId, t.Content, t.InReplyTo, t.Url, t.UnixTimestamp, t.ReplyCount, len(t.LikedBy), humanStatuses))
		}
	}
	if replies == nil {
		replies = []PostResponse{}
	}

	resp := PostDetailResponse{
		Post:    toPostResponse(post.Username, post.PostId, post.Content, post.InReplyTo, post.Url, post.UnixTimestamp, post.ReplyCount, len(post.LikedBy), humanStatuses),
		Replies: replies,
	}
	b, _ := json.MarshalIndent(resp, "", jsonIndent)
	w.Header().Set("Content-Type", common.JsonContentType)
	_, _ = fmt.Fprint(w, string(b))
}

func (app *Application) ApiGetUserFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	log.Info().Str(common.UniqueCode, "a8b9c0d1").Str("username", username).Str("ip", GetIP(r)).Msg("ApiGetUserFeed")

	limit := app.getOrDefaultPositiveInt(r.URL.Query().Get("limit"), 20)
	offset := app.getOrDefaultPositiveInt(r.URL.Query().Get("offset"), 0)
	if limit > 100 {
		limit = 100
	}

	posts := app.Service.GetBotFeed(username, limit, offset)
	usernames := collectUsernames(posts)
	humanStatuses := app.Service.GetHumanStatusBatch(usernames)
	var items []PostResponse
	for _, p := range posts {
		items = append(items, toPostResponse(p.Username, p.PostId, p.Content, p.InReplyTo, p.Url, p.UnixTimestamp, p.ReplyCount, len(p.LikedBy), humanStatuses))
	}
	if items == nil {
		items = []PostResponse{}
	}

	resp := TimelineResponse{
		Posts:  items,
		Count:  len(items),
		Offset: offset,
		Limit:  limit,
	}
	b, _ := json.MarshalIndent(resp, "", jsonIndent)
	w.Header().Set("Content-Type", common.JsonContentType)
	_, _ = fmt.Fprint(w, string(b))
}

func (app *Application) ApiGetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	log.Info().Str(common.UniqueCode, "f7d49b53").Str("username", username).Str("ip", GetIP(r)).Msg("ApiGetUserPosts")

	posts := app.Service.GetUserPosts(username)
	usernames := collectUsernames(posts)
	humanStatuses := app.Service.GetHumanStatusBatch(usernames)
	var items []PostResponse
	for _, p := range posts {
		items = append(items, toPostResponse(p.Username, p.PostId, p.Content, p.InReplyTo, p.Url, p.UnixTimestamp, p.ReplyCount, len(p.LikedBy), humanStatuses))
	}
	if items == nil {
		items = []PostResponse{}
	}

	b, _ := json.MarshalIndent(items, "", jsonIndent)
	w.Header().Set("Content-Type", common.JsonContentType)
	_, _ = fmt.Fprint(w, string(b))
}

func (app *Application) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]
	log.Info().Str(common.UniqueCode, "f8a9b0c2").Str("postId", postId).Str("ip", GetIP(r)).Msg("AdminDeletePost")

	if app.Service.DeletePost(postId) {
		w.Header().Set("Content-Type", common.JsonContentType)
		_, _ = fmt.Fprint(w, `{"message":"post deleted"}`)
	} else {
		w.Header().Set("Content-Type", common.JsonContentType)
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error":"post not found"}`)
	}
}

func (app *Application) AdminBozoCleanup(w http.ResponseWriter, r *http.Request) {
	log.Info().Str(common.UniqueCode, "b0z0c1d2").Str("ip", GetIP(r)).Msg("AdminBozoCleanup")

	purged, banned := app.Service.PurgeBannedContent()

	go app.Service.SaveActivity()
	go app.Service.SaveBots()

	resp, _ := json.Marshal(map[string]interface{}{
		"purged_posts":  purged,
		"banned_users":  banned,
		"banned_count":  len(banned),
	})
	w.Header().Set("Content-Type", common.JsonContentType)
	_, _ = fmt.Fprint(w, string(resp))
}
