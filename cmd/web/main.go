package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xrinful/Zenq/internal/requester/flare"
)

func main() {
	solver := flare.New("http://0.0.0.0:8191")
	result, err := solver.GetCookies(
		context.Background(),
		"https://lek-manga.net/manga/the-oracle-of-the-villainous-baby/23/",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
