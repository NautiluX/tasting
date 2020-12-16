package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sdomino/scribble"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Config Config
	Db     *scribble.Driver
}

type Model struct {
	Items     []Item     `json:"items"`
	Questions []Question `json:"questions"`
	Texts     Texts      `json:"texts"`
}

type Config struct {
	Items     []Item     `yaml:"items"`
	Questions []Question `yaml:"questions"`
	Texts     Texts      `yaml:"texts"`
}

type Texts struct {
	Title string `yaml:"title" json:"title"`
	Jumbo string `yaml:"jumbo" json:"jumbo"`
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

	setupDb()
	server := NewServer()
	fs := http.FileServer(http.Dir("./content"))
	http.HandleFunc("/model", server.ModelHandler)
	http.HandleFunc("/rating", server.RatingHandler)
	http.HandleFunc("/comment", server.CommentHandler)
	http.Handle("/", fs)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func setupDb() {
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

	db, err := scribble.New("db", nil)
	if err != nil {
		panic(err)
	}

	s := Server{config, db}
	return &s
}

func (s *Server) readModel() *Model {
	items := []Item{}
	for _, item := range s.Config.Items {
		newItem := Item{}
		err := s.Db.Read("item", item.Key, &newItem)
		if err != nil {
			err := s.Db.Write("item", item.Key, &item)
			if err != nil {
				panic(err)
			}
			newItem = item
		}
		items = append(items, newItem)
	}
	model := Model{items, s.Config.Questions, s.Config.Texts}
	model.update()
	return &model
}

func (s *Server) ModelHandler(res http.ResponseWriter, req *http.Request) {
	model := s.readModel()
	result, err := json.Marshal(model)
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
			m.Items[i].Comments = []Comment{}
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
	model := s.readModel()
	resultString := ""
	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)

		var comment Comment
		err := decoder.Decode(&comment)

		if err != nil {
			resultString = fmt.Sprintf("%v", err)
			s.InternalError(res, resultString)
			return
		}
		if comment.Text == "" || comment.Author == "" || comment.Item == "" {
			s.InternalError(res, "missing field")
			return
		}

		for i, item := range model.Items {
			if item.Key == comment.Item {
				model.Items[i].Comments = append(model.Items[i].Comments, comment)
				err := s.Db.Write("item", item.Key, model.Items[i])
				if err != nil {
					s.InternalErrorFromErr(res, err)
					return
				}
			}
		}
		s.Success(res)
		return
	}
	s.InternalError(res, resultString)
}

func (s *Server) InternalErrorFromErr(res http.ResponseWriter, err error) {
	errString := fmt.Sprintf("Error: %v", err)
	s.InternalError(res, errString)
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
	model := s.readModel()
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
		if vote.Rating < 1 || vote.Rating > 5 || vote.Author == "" || vote.Item == "" || vote.RatingNum < 0 || vote.RatingNum >= len(model.Questions) {
			s.InternalError(res, "missing field")
			return
		}

		for i, item := range model.Items {
			if item.Key == vote.Item {
				for v, existingVote := range model.Items[i].Rating[vote.RatingNum].Votes {
					if existingVote.Author == vote.Author {
						model.Items[i].Rating[vote.RatingNum].Votes[v] = vote
						err := s.Db.Write("item", item.Key, model.Items[i])
						if err != nil {
							s.InternalErrorFromErr(res, err)
							return
						}
						s.Success(res)
						return
					}
				}
				model.Items[i].Rating[vote.RatingNum].Votes = append(model.Items[i].Rating[vote.RatingNum].Votes, vote)
				err := s.Db.Write("item", item.Key, model.Items[i])
				if err != nil {
					s.InternalErrorFromErr(res, err)
					return
				}
			}
		}
		s.Success(res)
		return
	}
	s.InternalError(res, resultString)
}
