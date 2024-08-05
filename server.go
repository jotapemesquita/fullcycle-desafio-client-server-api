package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type CotacaoDolarResponse struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type CotacaoDolar struct {
	ID         string
	Bid        string
	CreateDate string
}

func NewCotacaoDolar(bid string, createDate string) *CotacaoDolar {
	return &CotacaoDolar{
		ID:         uuid.New().String(),
		Bid:        bid,
		CreateDate: createDate,
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	cotacao := BuscaCotacaoDolar()
	RegistraCotacaoDolar(cotacao)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println(cotacao)
	json.NewEncoder(w).Encode(cotacao.Bid)
}

func main() {
	setupDatabase()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func BuscaCotacaoDolar() CotacaoDolarResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Timeout ao buscar a cotação do dólar")
			return CotacaoDolarResponse{}
		}

		panic(err)
	}

	defer res.Body.Close()
	cotacao := []CotacaoDolarResponse{}
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		panic(err)
	}

	log.Println(cotacao)
	return cotacao[0]
}

func RegistraCotacaoDolar(cotacao CotacaoDolarResponse) {
	db := getConnection()
	defer db.Close()

	c := NewCotacaoDolar(cotacao.Bid, cotacao.CreateDate)
	stmt, err := db.Prepare("INSERT INTO cotacao (id, bid, create_date) VALUES (?, ?, ?)")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = stmt.ExecContext(ctx, c.ID, c.Bid, c.CreateDate)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Timeout ao inserir a cotação no banco de dados")
			return
		}

		panic(err)
	}
}

func setupDatabase() {
	db := getConnection()
	defer db.Close()

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS cotacao (id TEXT PRIMARY KEY, bid REAL, create_date TEXT)")
	if err != nil {
		panic(err)
	}
}

func getConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		panic(err)
	}

	return db
}
