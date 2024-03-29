package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Usdbrl struct {
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
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	log.Println("Request iniciada")

	res, err := requestAPI()
	if err != nil {
		return
	}

	insertCotacao(res)
	returnoJson(w, r, res)
	log.Println("Request Finalizada")

}

func requestAPI() (*Cotacao, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var cotacao Cotacao
	var err error

	select {

	case <-time.After(0 * time.Second):
		req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v", err)
			return nil, err
		}

		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v", err)
			return nil, err
		}

		var cotacao Cotacao

		err = json.Unmarshal(res, &cotacao)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
			return nil, err
		}

		return &cotacao, err

	case <-ctx.Done():
		fmt.Println("Timeout acima de 200 milesgundos")
		return &cotacao, err
	}

}

func abreConexao() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "../banco/cotacao.db")
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	return db, err
}

func insertCotacao(cotacao *Cotacao) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db, err := abreConexao()
	if err != nil {
		fmt.Println("Erro ao conectar ao banco de dados:", err)
		return err
	}
	defer db.Close()

	statement := "insert into cotacao(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	// Execute the database query with the custom context
	_, err = db.ExecContext(ctx, statement, cotacao.Usdbrl.Code, cotacao.Usdbrl.Codein, cotacao.Usdbrl.Name, cotacao.Usdbrl.High, cotacao.Usdbrl.Low, cotacao.Usdbrl.VarBid, cotacao.Usdbrl.PctChange, cotacao.Usdbrl.Bid, cotacao.Usdbrl.Ask, cotacao.Usdbrl.Timestamp, cotacao.Usdbrl.CreateDate)
	if err != nil {
		fmt.Println("Erro ao executar query:", err)
		return err
	}

	return nil
}

func returnoJson(w http.ResponseWriter, r *http.Request, cotacao *Cotacao) {

	data := make(map[string]string)
	if cotacao.Usdbrl.Bid != "" {
		data = map[string]string{
			"Bid": cotacao.Usdbrl.Bid,
		}
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
