package utils_general

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var done = make(chan error)
var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:      100,
		IdleConnTimeout:   90 * time.Second,
		DisableKeepAlives: false,
	},
}

// Función genérica para hacer una solicitud POST
func PostRequest(request interface{}, ip string, port int, endpoint string) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("error serializando la solicitud: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("error creando la solicitud POST: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("error haciendo la solicitud POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("error en la solicitud POST, status code: %d", resp.StatusCode)
	}
}

func PostRequestNoBloqueante(request interface{}, ip string, port int, endpoint string) {
	go func() {
		jsonData, err := json.Marshal(request)
		if err != nil {
			done <- fmt.Errorf("error serializando la solicitud: %v", err)
			return
		}

		url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			done <- fmt.Errorf("error creando la solicitud POST: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			done <- fmt.Errorf("error haciendo la solicitud POST: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			done <- fmt.Errorf("error en la solicitud POST, status code: %d", resp.StatusCode)
			return
		}

		done <- nil // Sin errores
	}()
}

// Función genérica para manejar solicitudes POST y decodificar el cuerpo de la solicitud
func HandlePostRequest(w http.ResponseWriter, r *http.Request, req interface{}) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		log.Printf("Método no permitido: %s\n", r.Method)
		return
	}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "JSON inválido en el cuerpo de la solicitud", http.StatusBadRequest)
		log.Printf("Error al decodificar JSON: %v\n", err)
		return
	}

	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Solicitud procesada exitosamente"))
}
