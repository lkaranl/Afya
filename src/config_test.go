package main

import (
	"testing"
	"time"
)

func TestLoadTimezoneDefault(t *testing.T) {
	loc, name := loadTimezone("")
	if name != defaultTimezone {
		t.Errorf("esperado nome padrão %s, obtido %s", defaultTimezone, name)
	}
	if loc == nil {
		t.Fatal("location não deveria ser nula")
	}
	if _, offset := time.Now().In(loc).Zone(); offset != -3*3600 {
		t.Errorf("esperado offset UTC-3 para o fuso padrão, obtido %d", offset)
	}
}

func TestLoadTimezoneRondonia(t *testing.T) {
	loc, name := loadTimezone("America/Porto_Velho")
	if name != "America/Porto_Velho" {
		t.Errorf("esperado America/Porto_Velho, obtido %s", name)
	}
	if _, offset := time.Now().In(loc).Zone(); offset != -4*3600 {
		t.Errorf("esperado offset UTC-4 para Rondônia, obtido %d", offset)
	}
}

func TestLoadTimezoneInvalid(t *testing.T) {
	loc, name := loadTimezone("Fuso/Inexistente")
	if name != defaultTimezone {
		t.Errorf("esperado fallback para %s, obtido %s", defaultTimezone, name)
	}
	if loc == nil {
		t.Fatal("location de fallback não deveria ser nula")
	}
}