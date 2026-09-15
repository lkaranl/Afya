package main

import (
	"testing"
)

func TestGetTeacherGuideAll(t *testing.T) {
	guide, err := GetTeacherGuide("", "")
	if err != nil {
		t.Fatalf("erro ao obter guia do docente: %v", err)
	}

	if len(guide.Sections) == 0 {
		t.Errorf("esperava seções no guia, retornou 0")
	}

	if len(guide.Checklist) == 0 {
		t.Errorf("esperava checklist no guia, retornou 0")
	}

	if len(guide.Troubleshooting) == 0 {
		t.Errorf("esperava troubleshooting no guia, retornou 0")
	}

	if guide.Markdown == "" {
		t.Errorf("esperava markdown formatado, retornou vazio")
	}
}

func TestGetTeacherGuideTopicFilter(t *testing.T) {
	guide, err := GetTeacherGuide("avaliacoes", "")
	if err != nil {
		t.Fatalf("erro ao obter guia: %v", err)
	}

	if len(guide.Sections) != 1 {
		t.Errorf("esperava 1 seção para 'avaliacoes', retornou %d", len(guide.Sections))
	}

	if guide.Sections[0].ID != "avaliacoes" {
		t.Errorf("esperava ID 'avaliacoes', retornou %s", guide.Sections[0].ID)
	}
}

func TestGetTeacherGuideSearch(t *testing.T) {
	guide, err := GetTeacherGuide("", "speedgrader")
	if err != nil {
		t.Fatalf("erro ao pesquisar guia: %v", err)
	}

	if len(guide.Sections) == 0 {
		t.Errorf("esperava encontrar seções com 'speedgrader'")
	}
}
