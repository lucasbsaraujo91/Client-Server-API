package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	select {

	case <-time.After(0 * time.Second):

		req, err := http.Get("http://localhost:8080/cotacao")

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao requisitar dados da api: %v\n", err)
		}

		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		}

		var data Cotacao
		err = json.Unmarshal(res, &data)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
		}

		file, err := os.Create("cotacao.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
		}

		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("Dólar: %s", data.Bid))

		fmt.Println(data)

	case <-ctx.Done():
		fmt.Println("Timeout acima de 300 milesegundos")
	}

}
