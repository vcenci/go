package main

import "net/http"
import "fmt"
import "encoding/json"

const CAPACITY = 100

type Message struct {
    ID   int
    Body string
}

// O problema aqui é a capacidade. Se atingir a capacidade, as requisições de enfileirar vão ficar bloqueadas até que haja espaço na fila.
// Se não definirmos capacidade, a criação de mensagens fica bloqueada até que haja um consumidor para ler.
// Se quisermos ter filas que crescem indefinidamente precisaremos criar e gerenciar a estrutura na memória.
func main() {
	fmt.Println("Iniciando servidor na porta 8080")

	queue := make(chan Message, CAPACITY)

	http.HandleFunc("/queue", func(w http.ResponseWriter, req *http.Request) {
		var message Message
		err := json.NewDecoder(req.Body).Decode(&message)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		if message.ID == 0 || message.Body == "" {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "'ID' ou 'Body' não informados no json", http.StatusBadRequest)
			return
		}

		queue <- message
	})

	http.HandleFunc("/consume", func(w http.ResponseWriter, req *http.Request) {
		message := <-queue
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", message)))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erro ao criar servidor na porta 8080: ", err)
	}
}