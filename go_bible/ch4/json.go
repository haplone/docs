package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Movie struct {
	Title  string
	Year   int  `json:"released"`
	Color  bool `json:"color,omitempty"`
	Actors []string
}

var movies = []Movie{
	{Title: "Casablanca", Year: 1942, Color: false,
		Actors: []string{"Humphrey Bogart", "Ingrid Bergman"}},
	{Title: "Cool Hand Luke", Year: 1967, Color: true,
		Actors: []string{"Paul Newman"}},
	{Title: "Bullitt", Year: 1968, Color: true,
		Actors: []string{"Steve McQueen", "Jacqueline Bisset"}},
}

func main() {
	data, err := json.Marshal(movies)
	if err != nil {
		log.Fatalf("JSON marshaling failed, %s", err)
	}
	fmt.Printf("%s\n", data)

	data2, err2 := json.MarshalIndent(movies, "", "	")
	if err2 != nil {
		log.Fatalf("json marshaling failed ,%s", err2)
	}
	fmt.Printf("%s \n", data2)

	var original []Movie
	err3 := json.Unmarshal(data2, &original)
	if err3 != nil {
		log.Fatalf("json unmarshaling failed %s", err3)
	}
	fmt.Println(original)

	var terms = []string{"repo:golang/go", "is:open"}
	result, err := SearchIssues(terms)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d issues:\n", result.TotalCount)
	fmt.Println(result)
	for _, item := range result.Items {
		fmt.Printf("#%-5d %9.9s %5.25s\n", item.Number, item.User.Login, item.Title)
	}

	report, err := template.New("report").Funcs(template.FuncMap{"daysAgo": daysAgo}).
		Parse(templ)

	if err != nil {
		log.Fatal(err)
	}
	if err := report.Execute(os.Stdout, result); err != nil {
		log.Fatal(err)
	}
}

func daysAgo(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}

const IssuesURL = "https://api.github.com/search/issues"

const templ = `{{.TotalCount}} issues:
{{range .Items}}------------------------------------
Number: {{.Number}}
User:   {{.User.Login}}
Title:  {{.Title |printf "%.64s"}}
Age:    {{.CreatedAt|daysAgo}} days
{{end}}`

type IssuesSearchResult struct {
	TotalCount int `json:"total_count"`
	Items      []*Issue
}

type Issue struct {
	Number    int
	HTMLURL   string `json:"html_url"`
	Title     string
	State     string
	User      *User
	CreatedAt time.Time `json:"created_at"`
	Body      string
}

type User struct {
	Login   string
	HTMLURL string `json:"html_url"`
}

func SearchIssues(terms []string) (*IssuesSearchResult, error) {
	q := url.QueryEscape(strings.Join(terms, " "))
	resp, err := http.Get(IssuesURL + "?q=" + q)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("search query failed :%s", resp.Status)
	}

	var result IssuesSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return nil, err
	}

	resp.Body.Close()
	return &result, nil
}
