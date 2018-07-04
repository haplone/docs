package main

/**
 * http://www.opscoder.info/boltdb_intro.html
 */
import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

func main() {
	db, err := bolt.Open("blog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("posts"))
		if err != nil {
			return err
		}
		return b.Put([]byte("2015-01-01"), []byte("my new year post"))
	})

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("posts"))
		v := b.Get([]byte("2015-01-01"))
		log.Println(string(v))
		return nil
	})

	post := &Post{
		Created: time.Now(),
		Title:   "my first post",
		Content: "hello bolt",
	}
	updatePost(db, post)
	readPost(db, post.Created)
}

type Post struct {
	Created time.Time
	Title   string
	Content string
}

func updatePost(db *bolt.DB, p *Post) {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("posts"))
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(p)
		if err != nil {
			return nil
		}
		return b.Put([]byte(p.Created.Format(time.RFC3339)), encoded)
	})
}

func readPost(db *bolt.DB, t time.Time) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("posts"))
		v := b.Get([]byte(t.Format(time.RFC3339)))
		log.Println(string(v))
		return nil
	})

}
