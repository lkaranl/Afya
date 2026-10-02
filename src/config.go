package main

import (
	"log"
	"os"
	"sync"
	"time"
)

// Nome padrão do fuso horário institucional, usado quando TIMEZONE não está
// configurada no .env (preserva o comportamento histórico do sistema: UTC-3).
const defaultTimezone = "America/Sao_Paulo"

var (
	tzOnce sync.Once
	tzLoc  *time.Location
	tzName string
)

// loadTimezone resolve um nome de fuso horário IANA (ex: "America/Porto_Velho")
// para um *time.Location, com fallback seguro para UTC-3 em caso de nome inválido.
func loadTimezone(name string) (*time.Location, string) {
	if name == "" {
		name = defaultTimezone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("⚠️ [TIMEZONE] Fuso horário inválido '%s' no .env; usando '%s' como fallback.", name, defaultTimezone)
		name = defaultTimezone
		loc, _ = time.LoadLocation(defaultTimezone)
		if loc == nil {
			loc = time.FixedZone("BRT", -3*60*60)
		}
	}
	return loc, name
}

// InstitutionTimezone retorna o fuso horário institucional configurado na
// variável de ambiente TIMEZONE (.env). O resultado é resolvido uma única vez
// (singleton) e reutilizado por todas as formatações de data do sistema.
func InstitutionTimezone() *time.Location {
	tzOnce.Do(func() {
		tzLoc, tzName = loadTimezone(os.Getenv("TIMEZONE"))
	})
	return tzLoc
}

// InstitutionTimezoneName retorna o nome IANA do fuso horário institucional
// (ex: "America/Porto_Velho" para Rondônia, UTC-4).
func InstitutionTimezoneName() string {
	InstitutionTimezone()
	return tzName
}