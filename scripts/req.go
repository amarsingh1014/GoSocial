package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)


type updatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func updatePost(postID int, p updatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://localhost:8080/v1/posts/%d", postID)

	b, _ := json.Marshal(p)

	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(b))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
}

func main() {

	var wg sync.WaitGroup

	postID := 8

	wg.Add(2)

	content := "This is a new content now"

	title := "This is a new title now"

	go updatePost(postID, updatePostPayload{Title: &title}, &wg)
	go updatePost(postID, updatePostPayload{Content: &content}, &wg)

	wg.Wait()

}