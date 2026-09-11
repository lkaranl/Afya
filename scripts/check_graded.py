#!/usr/bin/env python3
"""
Utilitário para Diagnóstico e Status de Correção de Atividades no Canvas LMS.
Exibe detalhadamente quais atividades já foram corrigidas, quais estão parcialmente
avaliadas e quais ainda possuem pendências de notas.

Uso:
  python3 scripts/check_graded.py                          # Lista disciplinas disponíveis
  python3 scripts/check_graded.py <course_id>               # Status de todas as atividades da disciplina
  python3 scripts/check_graded.py <course_id> <assign_id>   # Detalhes e lista de alunos pendentes/corrigidos
  python3 scripts/check_graded.py <course_id> --json        # Exporta relatório consolidado em JSON
"""

import sys
import os
import json
from datetime import datetime
from typing import Any, Dict, List, Optional

# Permite importar canvas_cli de scripts/ ou do diretório atual
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
if SCRIPT_DIR not in sys.path:
    sys.path.insert(0, SCRIPT_DIR)

from canvas_cli import (
    list_courses,
    list_assignments,
    get_assignment,
    get_submissions
)


def format_iso_date(iso_str: Optional[str]) -> str:
    """Formata data ISO UTC para formato amigável no horário brasileiro."""
    if not iso_str:
        return "Sem prazo definido"
    try:
        # Se contiver 'Z', substitui por +00:00
        dt = datetime.fromisoformat(iso_str.replace("Z", "+00:00"))
        # Converte para UTC-3 (Horário de Brasília) aproximadamente
        from datetime import timezone, timedelta
        br_tz = timezone(timedelta(hours=-3))
        dt_br = dt.astimezone(br_tz)
        return dt_br.strftime("%d/%m/%Y às %H:%M")
    except Exception:
        return iso_str


def analyze_assignment(course_id: str, assignment: Dict[str, Any]) -> Dict[str, Any]:
    """Analisa o status de correção de uma atividade específica."""
    aid = str(assignment.get("id"))
    name = assignment.get("name", "Sem nome")
    points = assignment.get("points_possible", 0)
    due_at = assignment.get("due_at")

    subs = get_submissions(course_id, aid, only_pending=False)

    total = len(subs)
    graded_list = []
    pending_list = []
    unsubmitted_list = []

    for s in subs:
        user_info = {
            "user_id": s.get("user_id"),
            "user_name": s.get("user_name", "Desconhecido"),
            "grade": s.get("grade"),
            "workflow_state": s.get("workflow_state"),
            "submitted_at": s.get("submitted_at")
        }
        
        # Considera corrigida se tiver nota atribuída ou estado "graded"
        if s.get("workflow_state") == "graded" or s.get("grade") is not None:
            graded_list.append(user_info)
        elif s.get("workflow_state") == "submitted" or s.get("submitted_at") is not None:
            pending_list.append(user_info)
        else:
            unsubmitted_list.append(user_info)

    graded_count = len(graded_list)
    pending_count = len(pending_list)
    unsubmitted_count = len(unsubmitted_list)

    if total > 0:
        percent_graded = (graded_count / total) * 100
    else:
        percent_graded = 0.0

    if pending_count == 0 and graded_count > 0:
        status_category = "CONCLUIDA"
    elif graded_count > 0 and pending_count > 0:
        status_category = "PARCIAL"
    elif graded_count == 0 and pending_count > 0:
        status_category = "PENDENTE"
    else:
        status_category = "SEM_ENTREGAS"

    return {
        "id": aid,
        "name": name,
        "points": points,
        "due_at": due_at,
        "due_formatted": format_iso_date(due_at),
        "status_category": status_category,
        "total_enrolled_submissions": total,
        "graded_count": graded_count,
        "pending_count": pending_count,
        "unsubmitted_count": unsubmitted_count,
        "percent_graded": round(percent_graded, 1),
        "graded_students": graded_list,
        "pending_students": pending_list,
        "unsubmitted_students": unsubmitted_list
    }


def print_course_status(course_id: str, as_json: bool = False):
    """Exibe o status de correção de todas as tarefas de uma disciplina."""
    assignments = list_assignments(course_id)
    if not assignments:
        print(f"Nenhuma atividade encontrada para o curso {course_id}.")
        return

    results = [analyze_assignment(course_id, a) for a in assignments]

    if as_json:
        print(json.dumps(results, indent=2, ensure_ascii=False))
        return

    print("=" * 80)
    print(f"📊 STATUS DE CORREÇÕES - DISCIPLINA {course_id}")
    print("=" * 80)

    concluidas = [r for r in results if r["status_category"] == "CONCLUIDA"]
    parciais = [r for r in results if r["status_category"] == "PARCIAL"]
    pendentes = [r for r in results if r["status_category"] == "PENDENTE"]
    vazias = [r for r in results if r["status_category"] == "SEM_ENTREGAS"]

    if concluidas:
        print("\n✅ ATIVIDADES 100% CORRIGIDAS:")
        for r in concluidas:
            print(f"  • [{r['id']}] {r['name']}")
            print(f"    Notas: {r['graded_count']}/{r['total_enrolled_submissions']} ({r['percent_graded']}%) | Pontos: {r['points']} | Prazo: {r['due_formatted']}")

    if parciais:
        print("\n⏳ ATIVIDADES PARCIALMENTE CORRIGIDAS (PENDÊNCIAS RESTANTES):")
        for r in parciais:
            print(f"  • [{r['id']}] {r['name']}")
            print(f"    Corrigidas: {r['graded_count']} | Aguardando correção: {r['pending_count']} | Total: {r['total_enrolled_submissions']} ({r['percent_graded']}%)")
            print(f"    Pontos: {r['points']} | Prazo: {r['due_formatted']}")

    if pendentes:
        print("\n🔴 ATIVIDADES COM CORREÇÃO TOTALMENTE PENDENTE:")
        for r in pendentes:
            print(f"  • [{r['id']}] {r['name']}")
            print(f"    Aguardando correção: {r['pending_count']} | Pontos: {r['points']} | Prazo: {r['due_formatted']}")

    if vazias:
        print("\n⚪ ATIVIDADES SEM SUBMISSÕES OU NÃO ENTREGUES:")
        for r in vazias:
            print(f"  • [{r['id']}] {r['name']} | Prazo: {r['due_formatted']}")

    print("\n" + "=" * 80)


def print_single_assignment(course_id: str, assignment_id: str, as_json: bool = False):
    """Exibe os detalhes aprofundados de uma única atividade."""
    assignment = get_assignment(course_id, assignment_id)
    result = analyze_assignment(course_id, assignment)

    if as_json:
        print(json.dumps(result, indent=2, ensure_ascii=False))
        return

    print("=" * 80)
    print(f"📌 DETALHES DA ATIVIDADE: {result['name']} (ID: {result['id']})")
    print(f"Prazo: {result['due_formatted']} | Pontuação Máxima: {result['points']}")
    print(f"Progresso: {result['graded_count']}/{result['total_enrolled_submissions']} avaliadas ({result['percent_graded']}%)")
    print("=" * 80)

    if result["pending_students"]:
        print(f"\n⏳ Alunos Aguardando Correção ({len(result['pending_students'])}):")
        for idx, s in enumerate(result["pending_students"], 1):
            sub_at = format_iso_date(s.get("submitted_at"))
            print(f"  {idx}. {s['user_name']} (ID: {s['user_id']}) - Entregue em: {sub_at}")

    if result["graded_students"]:
        print(f"\n✅ Alunos Já Corrigidos ({len(result['graded_students'])}):")
        for idx, s in enumerate(result["graded_students"], 1):
            print(f"  {idx}. {s['user_name']} (ID: {s['user_id']}) - Nota: {s['grade']}")

    print("\n" + "=" * 80)


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    as_json = "--json" in sys.argv

    if not args:
        print("Buscando turmas ativas do professor...")
        courses = list_courses()
        print("\n📚 Disciplinas Encontradas:")
        for c in courses:
            print(f"  • ID: {c.get('id')} | Código: {c.get('course_code')} | Nome: {c.get('name')}")
        print("\nPara ver o status das correções de uma disciplina, execute:")
        print("  python3 scripts/check_graded.py <course_id>")
        sys.exit(0)

    course_id = args[0]
    if len(args) >= 2:
        assignment_id = args[1]
        print_single_assignment(course_id, assignment_id, as_json=as_json)
    else:
        print_course_status(course_id, as_json=as_json)


if __name__ == "__main__":
    main()
