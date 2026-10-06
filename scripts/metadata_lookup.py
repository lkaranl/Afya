#!/usr/bin/env python3
"""
Módulo de Consulta Rápida aos Metadados do Semestre Vigente (2026.2).
Permite resolver instantaneamente IDs de turmas e alunos sem gastar chamadas de rede nem tokens.
"""

import os
import json
import unicodedata
from typing import Optional, Dict, Any, List

METADATA_FILE = os.path.join(os.path.dirname(__file__), '..', 'data', 'metadata', 'semestre_2026_2.json')
METADATA_FILE = os.path.abspath(METADATA_FILE)

_CACHED_DATA: Optional[Dict[str, Any]] = None

def _load_data() -> Dict[str, Any]:
    global _CACHED_DATA
    if _CACHED_DATA is None:
        if not os.path.exists(METADATA_FILE):
            raise FileNotFoundError(f"Arquivo de metadados não encontrado em {METADATA_FILE}. Execute scripts/sync_metadata.py primeiro.")
        with open(METADATA_FILE, 'r', encoding='utf-8') as f:
            _CACHED_DATA = json.load(f)
    return _CACHED_DATA

def normalize_text(text: str) -> str:
    if not text:
        return ""
    nfkd = unicodedata.normalize('NFKD', text)
    sem_acento = "".join([c for c in nfkd if not unicodedata.combining(c)])
    return sem_acento.strip().lower()

def get_course_id(periodo_ou_slug: str) -> Optional[str]:
    """Retorna o ID numérico do curso no Canvas a partir do período (ex: '2', '2p', '2º Período', '4º Período')."""
    d = _load_data()
    norm = normalize_text(periodo_ou_slug)
    
    # 2º período
    if any(k in norm for k in ['2', 'segundo', '2p', '2o']):
        return d["turmas"].get("2_periodo", {}).get("course_id")
    # 4º período
    if any(k in norm for k in ['4', 'quarto', '4p', '4o']):
        return d["turmas"].get("4_periodo", {}).get("course_id")
    
    # Busca direta por chave
    if periodo_ou_slug in d["turmas"]:
        return d["turmas"][periodo_ou_slug].get("course_id")
    return None

def find_student(nome_ou_termo: str, course_id: Optional[str] = None) -> List[Dict[str, Any]]:
    """Busca alunos pelo nome ou parte do nome sem se preocupar com maiúsculas ou acentos."""
    d = _load_data()
    query = normalize_text(nome_ou_termo)
    results = []

    for chave_busca, aluno_info in d.get("indice_rapido", {}).get("alunos_para_curso", {}).items():
        if query in chave_busca:
            if course_id and str(aluno_info.get("curso_id")) != str(course_id):
                continue
            results.append(aluno_info)
    return results

def get_all_students(periodo_ou_slug: str) -> List[Dict[str, Any]]:
    """Retorna a lista completa de alunos de uma turma."""
    d = _load_data()
    cid = get_course_id(periodo_ou_slug)
    slug = d.get("indice_rapido", {}).get("por_curso_id", {}).get(cid)
    if slug and slug in d["turmas"]:
        return d["turmas"][slug].get("alunos", [])
    return []

if __name__ == '__main__':
    print("Teste de Lookup Local:")
    print("ID 2º Período:", get_course_id("2º Período"))
    print("ID 4º Período:", get_course_id("4º Período"))
    print("Busca 'wesley':", find_student("wesley"))
    print("Busca 'thiago':", find_student("thiago"))
