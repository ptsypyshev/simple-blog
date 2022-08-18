package main

import (
	"github.com/ptsypyshev/simple-blog/internal/blog"
	"log"
)

func main() {
	a := blog.App{}
	if closer, err := a.Init(); err != nil {
		log.Fatal(err)
	} else {
		defer closer.Close()
	}

	if err := a.Serve(); err != nil {
		log.Fatal(err)
	}
}
