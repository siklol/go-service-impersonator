package main

import (
	"encoding/json"
	"github.com/siklol/go-service-impersonator/docker"
)

func main() {
	impersonator, err := docker.NewImpersonator()

	if err != nil {
		panic(err)
	}

	var j map[string]interface{}
	json.Unmarshal([]byte(`{
  "posts": [
    { "id": 1, "body": "foo" },
    { "id": 2, "body": "bar" }
  ],
  "comments": [
    { "id": 1, "body": "baz", "postId": 1 },
    { "id": 2, "body": "qux", "postId": 2 }
  ]
}`), &j)
	if err := impersonator.RESTService(j); err != nil {
		panic(err)
	}
}
