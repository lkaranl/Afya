package main

import (
	"testing"
)

func TestPlagiarismAlgorithm(t *testing.T) {
	// Código Original
	codeOriginal := `
    #include <stdio.h>
    int busca_binaria(int vet[], int tam, int chave) {
        int inicio = 0;
        int fim = tam - 1;
        while (inicio <= fim) {
            int meio = (inicio + fim) / 2;
            if (vet[meio] == chave) {
                return meio;
            } else if (vet[meio] < chave) {
                inicio = meio + 1;
            } else {
                fim = meio - 1;
            }
        }
        return -1;
    }
    `

	// Código Plagiado com renomeação de variáveis e troca de comentários
	codePlagiarizedRenamed := `
    // Minha funcao de busca
    #include <stdio.h>
    int procurar_elemento(int arr[], int n, int valor) {
        int esq = 0;
        int dir = n - 1;
        while (esq <= dir) {
            int centro = (esq + dir) / 2;
            if (arr[centro] == valor) {
                return centro;
            } else if (arr[centro] < valor) {
                esq = centro + 1;
            } else {
                dir = centro - 1;
            }
        }
        return -1;
    }
    `

	// Código Diferente (Bubble Sort)
	codeDifferent := `
    #include <stdio.h>
    void ordenar_bolha(int lista[], int tamanho) {
        for (int i = 0; i < tamanho - 1; i++) {
            for (int j = 0; j < tamanho - i - 1; j++) {
                if (lista[j] > lista[j+1]) {
                    int temp = lista[j];
                    lista[j] = lista[j+1];
                    lista[j+1] = temp;
                }
            }
        }
    }
    `

	tokOrigCanon := tokenizeAndCanonicalize(codeOriginal, true)
	tokPlagCanon := tokenizeAndCanonicalize(codePlagiarizedRenamed, true)
	tokDiffCanon := tokenizeAndCanonicalize(codeDifferent, true)

	ngOrig := generateNGrams(tokOrigCanon, 4)
	ngPlag := generateNGrams(tokPlagCanon, 4)
	ngDiff := generateNGrams(tokDiffCanon, 4)

	simPlag, _ := calculateSimilarity(ngOrig, ngPlag)
	simDiff, _ := calculateSimilarity(ngOrig, ngDiff)

	t.Logf("Similaridade com cópia mascarada (variáveis trocadas): %.2f%%", simPlag)
	t.Logf("Similaridade com código diferente (Bubble Sort): %.2f%%", simDiff)

	if simPlag < 80.0 {
		t.Fatalf("Esperado similaridade alta (> 80%%) para código com variáveis trocadas, obtido: %.2f%%", simPlag)
	}

	if simDiff > 40.0 {
		t.Fatalf("Esperado similaridade baixa (< 40%%) para algoritmo completamente diferente, obtido: %.2f%%", simDiff)
	}
}
