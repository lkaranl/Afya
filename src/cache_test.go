package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryCacheBasic(t *testing.T) {
	cache := NewMemoryCache(100 * time.Millisecond)

	cache.Set("turma:1", "Estrutura de Dados", 0)

	val, found := cache.Get("turma:1")
	if !found || val.(string) != "Estrutura de Dados" {
		t.Fatalf("Esperava encontrar 'Estrutura de Dados', obteve %v (found: %v)", val, found)
	}

	// Aguarda expiração do TTL
	time.Sleep(150 * time.Millisecond)
	_, foundAfter := cache.Get("turma:1")
	if foundAfter {
		t.Fatalf("Esperava que o item tivesse expirado após o TTL")
	}

	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("Estatísticas incorretas: %+v", stats)
	}
}

func TestMemoryCacheDeletePrefix(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)

	cache.Set("assignments:101:task1", "Tarefa 1", 0)
	cache.Set("assignments:101:task2", "Tarefa 2", 0)
	cache.Set("assignments:202:task1", "Outra turma", 0)

	removed := cache.DeletePrefix("assignments:101")
	if removed != 2 {
		t.Errorf("Esperava remover 2 itens, removeu %d", removed)
	}

	if _, found := cache.Get("assignments:101:task1"); found {
		t.Errorf("Item não deveria existir após DeletePrefix")
	}
	if _, found := cache.Get("assignments:202:task1"); !found {
		t.Errorf("Item de outra turma deveria ter sido preservado")
	}
}

func TestMemoryCacheLatencySub50ms(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	data := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]any{
			"id":   fmt.Sprintf("student_%d", i),
			"name": fmt.Sprintf("Aluno %d", i),
		}
	}

	cache.Set("students:162263", data, 0)

	start := time.Now()
	val, found := cache.Get("students:162263")
	elapsed := time.Since(start)

	if !found {
		t.Fatalf("Esperava encontrar chave no cache")
	}
	if len(val.([]map[string]any)) != 100 {
		t.Fatalf("Tamanho inesperado do array")
	}

	// A latência deve ser muito menor que 50ms (geralmente < 100 microsegundos)
	if elapsed > 50*time.Millisecond {
		t.Fatalf("Latência excedeu 50ms: %v", elapsed)
	}
	t.Logf("Latência de recuperação do cache em memória: %v (< 50ms garantido)", elapsed)
}

func TestMemoryCacheConcurrency(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", id%5)
			cache.Set(key, id, 0)
			cache.Get(key)
			if id%10 == 0 {
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}
