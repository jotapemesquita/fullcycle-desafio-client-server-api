package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Timeout ao buscar a cotação do dólar")
			return
		}

		panic(err)
	}

	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}
