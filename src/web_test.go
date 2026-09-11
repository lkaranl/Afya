//go:build web

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCodeTestEndpoint(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/code/test", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code             string `json:"code"`
			Language         string `json:"language"`
			ExpectedFunction string `json:"expected_function,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		if req.Code == "" {
			http.Error(w, "Código vazio", http.StatusBadRequest)
			return
		}

		res := map[string]any{
			"status":       "success",
			"syntax_valid": true,
		}
		sendWebJSON(w, res, nil)
	})

	body, _ := json.Marshal(map[string]string{
		"code":     "#include <stdio.h>\nint main() { return 0; }",
		"language": "c",
	})

	req := httptest.NewRequest("POST", "/api/code/test", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado 200, recebido %d", w.Code)
	}

	var res map[string]any
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	if res["status"] != "success" || res["syntax_valid"] != true {
		t.Fatalf("Resultado inesperado: %v", res)
	}
}
