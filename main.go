package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type View struct {
	Tweets []twitter.Tweet
}
type keys struct {
	Consumerkey    string
	Consumersecret string
	Token          string
	Tokensecret    string
}

var (
	tokens = keys{}
	client *twitter.Client
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index", 301)
	})

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, r)
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		q := r.PostFormValue("q")

		consumerKey := r.PostFormValue("consumer-key")
		consumerSecret := r.PostFormValue("consumer-secret")
		token := r.PostFormValue("token")
		tokenSecret := r.PostFormValue("token-secret")

		tokens.setKeys(consumerKey, consumerSecret, token, tokenSecret)
		client = auth()

		var view View
		view.Tweets = findTheSearchedWords(q)
		t, _ := template.ParseFiles("list.html")
		t.Execute(w, view)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

		_, err := deleteTweet(id)

		if err != nil {
			log.Println(err)
		}
	})

	http.ListenAndServe(":8080", nil)

}

func (t *keys) setKeys(consumerKey string, consumerSecret string, token string, tokenSecret string) {
	tokens.Consumerkey = consumerKey
	tokens.Consumersecret = consumerSecret
	tokens.Token = token
	tokens.Tokensecret = tokenSecret
}

func auth() *twitter.Client {
	config := oauth1.NewConfig(tokens.Consumerkey, tokens.Consumersecret)
	token := oauth1.NewToken(tokens.Token, tokens.Tokensecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	return client
}

//found wanted word in user tweets
func findTheSearchedWords(q string) []twitter.Tweet {

	userTimeLine, err := getTweets(client, 0)

	if err != nil {
		return []twitter.Tweet{}
	}

	allTweets := userTimeLine
	oldestTweet := userTimeLine[len(allTweets)-1].ID

	for {
		userTimeLine, err = getTweets(client, oldestTweet)

		if err != nil {
			break
		}

		if oldestTweet == userTimeLine[len(userTimeLine)-1].ID {
			break
		}

		for i := range userTimeLine {
			allTweets = append(allTweets, userTimeLine[i])
		}

		oldestTweet = userTimeLine[len(userTimeLine)-1].ID

	}

	var tweets []twitter.Tweet

	for i := range allTweets {
		tweet := allTweets[i].Text
		if strings.Contains(tweet, q) {
			tweets = append(tweets, allTweets[i])
		}
	}
	return tweets
}

//get all user tweets from twitter
func getTweets(client *twitter.Client, oldestTweet int64) (tweets []twitter.Tweet, err error) {

	userTimeLine, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		Count: 200, //200 is the maximum allowed count from twitter api
		MaxID: oldestTweet})

	if err != nil {
		return userTimeLine, err
	}

	return userTimeLine, nil
}

//delete selected tweet
func deleteTweet(id int64) (status bool, err error) {

	_, _, e := client.Statuses.Destroy(id, &twitter.StatusDestroyParams{})

	if e != nil {
		return false, e
	}
	return true, nil
}

//delete all found tweets
func deleteAllTweets(tweets []twitter.Tweet, q string) {
	for i := range tweets {
		tweet := tweets[i].Text

		if strings.Contains(tweet, q) {
			_, _, err := client.Statuses.Destroy(tweets[i].ID, &twitter.StatusDestroyParams{})
			if err != nil {
				log.Println(err)
				break
			}
			log.Printf("%s is deleted!", tweet)
		}
	}
}
