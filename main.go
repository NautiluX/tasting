package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Config Config
	Model  *Model
}

type Model struct {
	Items     []Item     `json:"items"`
	Ratings   []Rating   `json:"ratings"`
	Questions []Question `json:"questions"`
}

type Config struct {
	Items     []Item     `yaml:"items"`
	Questions []Question `yaml:"questions"`
}

type Question struct {
	Name string `yaml:"name" json:"name"`
	Key  string `yaml:"key" json:"key"`
}

type Item struct {
	Name     string    `yaml:"name" json:"name"`
	Key      string    `yaml:"key" json:"key"`
	Comments []Comment `json:"comments"`
	Rating   []Rating  `json:"rating"`
}

type Comment struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	Item   string `json:"item"`
}

type Rating struct {
	Question  Question `json:"question"`
	AvgRating float64  `json:"avgRating"`
	Votes     []Vote   `json:"rating"`
}

type Vote struct {
	Author    string  `json:"author"`
	Rating    float64 `json:"rating"`
	RatingNum int     `json:"ratingNum"`
	Item      string  `json:"item"`
}

func main() {
	server := NewServer()
	fs := http.FileServer(http.Dir("./content"))
	http.HandleFunc("/model", server.ModelHandler)
	http.HandleFunc("/rating", server.RatingHandler)
	http.HandleFunc("/comment", server.CommentHandler)
	http.Handle("/", fs)
	http.ListenAndServe(":8080", nil)
}

func NewServer() *Server {
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	s := Server{config, &Model{config.Items, []Rating{}, config.Questions}}
	return &s
}

func (s *Server) ModelHandler(res http.ResponseWriter, req *http.Request) {
	s.Model.update()
	result, err := json.Marshal(s.Model)
	if err != nil {
		resultString := fmt.Sprintf("%v", err)
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(resultString))
		return
	}
	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write(result)
}

func (m *Model) update() {
	for i, item := range m.Items {
		if item.Rating == nil || len(item.Rating) != len(m.Questions) {
			m.Items[i].Rating = []Rating{}
			for _, question := range m.Questions {
				m.Items[i].Rating = append(m.Items[i].Rating, Rating{question, 0, []Vote{}})
			}
		}
		if item.Comments == nil {
			m.Items[i].Comments = []Comment{Comment{"Alice", "Hello World", ""}, Comment{"Bob", "Greatest beer ever!", ""}}
		}
		for r, rating := range item.Rating {
			var sum float64 = 0
			for _, vote := range rating.Votes {
				sum += vote.Rating
			}
			if len(rating.Votes) > 0 {
				m.Items[i].Rating[r].AvgRating = sum / float64(len(rating.Votes))
			}
		}
	}
}

func (s *Server) CommentHandler(res http.ResponseWriter, req *http.Request) {
	resultString := ""
	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)

		var comment Comment
		err := decoder.Decode(&comment)
		//comment.Item = html.EscapeString(comment.Item)
		//comment.Author = html.EscapeString(comment.Author)
		//comment.Text = html.EscapeString(comment.Text)

		if err != nil {
			resultString = fmt.Sprintf("%v", err)
			s.InternalError(res, resultString)
			return
		}
		if comment.Text == "" || comment.Author == "" || comment.Item == "" {
			s.InternalError(res, "missing field")
			return
		}

		for i, item := range s.Model.Items {
			if item.Key == comment.Item {
				s.Model.Items[i].Comments = append(s.Model.Items[i].Comments, comment)
			}
		}
		s.Success(res)
		return
	}
	s.InternalError(res, resultString)
}

func (s *Server) InternalError(res http.ResponseWriter, body string) {
	res.WriteHeader(http.StatusInternalServerError)
	_, _ = res.Write([]byte(body))
}

func (s *Server) Success(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write([]byte("{\"result\":\"success\"}"))
}

func (s *Server) RatingHandler(res http.ResponseWriter, req *http.Request) {
	resultString := ""
	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)

		var vote Vote
		err := decoder.Decode(&vote)

		if err != nil {
			resultString = fmt.Sprintf("%v", err)
			s.InternalError(res, resultString)
			return
		}
		if vote.Rating < 1 || vote.Rating > 5 || vote.Author == "" || vote.Item == "" || vote.RatingNum < 0 || vote.RatingNum >= len(s.Model.Questions) {
			s.InternalError(res, "missing field")
			return
		}

		for i, item := range s.Model.Items {
			if item.Key == vote.Item {
				for v, existingVote := range s.Model.Items[i].Rating[vote.RatingNum].Votes {
					if existingVote.Author == vote.Author {
						s.Model.Items[i].Rating[vote.RatingNum].Votes[v] = vote
						s.Success(res)
						return
					}
				}
				s.Model.Items[i].Rating[vote.RatingNum].Votes = append(s.Model.Items[i].Rating[vote.RatingNum].Votes, vote)
			}
		}
		s.Success(res)
		return
	}
	s.InternalError(res, resultString)
}
