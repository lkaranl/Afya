#!/usr/bin/env python3
"""
Sincronizador e Gerador de Metadados Locais do Canvas LMS (Semestre Vigente).
Salva turmas, tarefas, módulos e alunos em data/metadata/semestre_2026_2.json
para consulta ultrarrápida com custo zero de tokens e de rede.
"""

import os
import sys
import json
import unicodedata
from datetime import datetime

# Garantir import do canvas_cli
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'scripts'))
try:
    from canvas_cli import call_mcp
except ImportError:
    from scripts.canvas_cli import call_mcp

def normalize_text(text: str) -> str:
    """Normaliza texto removendo acentos e convertendo para minúsculas para busca flexível."""
    if not text:
        return ""
    nfkd = unicodedata.normalize('NFKD', text)
    sem_acento = "".join([c for c in nfkd if not unicodedata.combining(c)])
    return sem_acento.strip().lower()

def sync_active_term():
    print("Iniciando sincronização de metadados das disciplinas de 2026.2...")
    courses = call_mcp('canvas_list_courses', {'term_filter': 'current'})
    if not courses:
        print("Nenhuma turma ativa retornada.")
        return

    data = {
        "semestre": "2026/2",
        "gerado_em": datetime.now().strftime("%d/%m/%Y às %H:%M:%S"),
        "ano": 2026,
        "semestre_numero": 2,
        "resumo": {
            "total_turmas": len(courses),
            "turmas_ativas": []
        },
        "turmas": {},
        "indice_rapido": {
            "por_periodo": {},
            "por_curso_id": {},
            "alunos_para_curso": {}
        }
    }

    for c in courses:
        cid = str(c.get('id'))
        nome_completo = c.get('name', '')
        clean_name = c.get('clean_name', nome_completo)
        period = c.get('period', '')

        # Definir chave amigável precisa sem colisão
        if "4º" in period or "CDC.N.4" in nome_completo:
            slug_chave = "4_periodo"
        elif "2º" in period or "CDC.N.2" in nome_completo:
            slug_chave = "2_periodo"
        else:
            slug_chave = f"curso_{cid}"

        print(f"Processando turma: {clean_name} ({period}) - ID {cid}...")

        # 1. Alunos
        students = call_mcp('canvas_list_students', {'course_id': cid})
        alunos_lista = []
        mapa_alunos_id = {}
        for s in students:
            uid = str(s.get('id'))
            nome = s.get('name', '').strip()
            login = s.get('login_id', '')
            email = s.get('email', '')
            norm = normalize_text(nome)

            item_aluno = {
                "id": uid,
                "nome": nome,
                "nome_busca": norm,
                "email": email,
                "matricula": login
            }
            alunos_lista.append(item_aluno)
            mapa_alunos_id[uid] = nome
            data["indice_rapido"]["alunos_para_curso"][norm] = {
                "id": uid,
                "nome": nome,
                "curso_id": cid,
                "periodo": period
            }

        # 2. Módulos
        modulos = call_mcp('canvas_list_modules', {'course_id': cid})
        modulos_simplificados = []
        for m in (modulos or []):
            modulos_simplificados.append({
                "id": str(m.get('id')),
                "nome": m.get('name'),
                "posicao": m.get('position'),
                "publicado": m.get('published', True),
                "itens_count": m.get('items_count', 0)
            })

        # 3. Tarefas
        status_tarefas = call_mcp('canvas_get_grading_status', {'course_id': cid})
        tarefas_simplificadas = []
        for t in (status_tarefas.get('assignments', []) if isinstance(status_tarefas, dict) else []):
            tarefas_simplificadas.append({
                "id": str(t.get('assignment_id')),
                "nome": t.get('name'),
                "pontos": t.get('points_possible'),
                "pendentes_correcao": t.get('needs_grading_count', 0)
            })

        turma_info = {
            "course_id": cid,
            "chave": slug_chave,
            "periodo": period,
            "nome_oficial": nome_completo,
            "nome_didatico": clean_name,
            "total_alunos": len(alunos_lista),
            "alunos": alunos_lista,
            "mapa_alunos_id": mapa_alunos_id,
            "modulos": modulos_simplificados,
            "tarefas": tarefas_simplificadas
        }

        data["turmas"][slug_chave] = turma_info
        data["indice_rapido"]["por_periodo"][period] = cid
        data["indice_rapido"]["por_curso_id"][cid] = slug_chave
        data["resumo"]["turmas_ativas"].append({
            "chave": slug_chave,
            "periodo": period,
            "course_id": cid,
            "total_alunos": len(alunos_lista)
        })

    out_path = os.path.join(os.path.dirname(__file__), '..', 'data', 'metadata', 'semestre_2026_2.json')
    out_path = os.path.abspath(out_path)
    os.makedirs(os.path.dirname(out_path), exist_ok=True)

    with open(out_path, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

    print(f"\n✅ Metadados sincronizados e salvos com sucesso em: {out_path}")
    print(f"Total de turmas: {len(data['turmas'])}")
    for k, v in data["turmas"].items():
        print(f"  • {v['nome_didatico']} ({v['periodo']}): {v['total_alunos']} alunos matriculados, {len(v['modulos'])} módulos, {len(v['tarefas'])} tarefas cadastradas.")

if __name__ == '__main__':
    sync_active_term()
