#!/usr/bin/env python3
"""
Utilitário de Integração Canvas LMS & MCP Client
Facilita chamadas às ferramentas MCP do binário ./afya-canvas e tratamento de dados pedagógicos.
"""

import sys
import os
import json
import subprocess
import html
import re
import shutil
from typing import Any, Dict, List, Optional

# Localização padrão do binário MCP
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MCP_BINARY_PATH = os.path.join(PROJECT_ROOT, "afya-canvas")


def call_mcp(tool_name: str, arguments: Optional[Dict[str, Any]] = None) -> Any:
    """
    Executa uma ferramenta MCP no servidor Go via stdio e retorna o resultado parseado.
    """
    if not os.path.exists(MCP_BINARY_PATH):
        raise FileNotFoundError(f"Binário MCP não encontrado em: {MCP_BINARY_PATH}")

    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments or {}
        }
    }

    rpc_input = json.dumps(payload) + "\n"

    proc = subprocess.Popen(
        [MCP_BINARY_PATH],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=PROJECT_ROOT
    )

    stdout, stderr = proc.communicate(input=rpc_input)

    if proc.returncode != 0:
        raise RuntimeError(f"Falha na execução do MCP ({proc.returncode}): {stderr}")

    for line in stdout.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        try:
            resp = json.loads(line)
            if "error" in resp and resp["error"]:
                raise RuntimeError(f"Erro retornado pelo MCP: {resp['error']}")
            
            if "result" in resp and "content" in resp["result"]:
                content_list = resp["result"]["content"]
                if content_list and len(content_list) > 0:
                    text_data = content_list[0].get("text", "")
                    try:
                        return json.loads(text_data)
                    except json.JSONDecodeError:
                        return text_data
            return resp.get("result")
        except json.JSONDecodeError:
            continue

    raise RuntimeError(f"Não foi possível decodificar resposta JSON-RPC válida. Saída recebida: {stdout}")


def clean_canvas_html(raw_html: str) -> str:
    """
    Higieniza o HTML enviado pelo Canvas mantendo operadores de código (<, <=, >, &&)
    e quebras de linha corretas, evitando truncamento acidental de código-fonte.
    """
    if not raw_html:
        return ""

    # Normalizar quebras de tags comuns de bloco
    text = raw_html.replace("<br>", "\n").replace("<br/>", "\n").replace("<br />", "\n")
    text = re.sub(r"</(p|div|li|h[1-6]|tr)>", "\n", text, flags=re.IGNORECASE)

    # Remover tags HTML sem devorar operadores de comparação
    text = re.sub(r"<[^>]+>", "", text)

    # Desescapar entidades HTML
    text = html.unescape(text)

    # Normalizar espaços não-quebráveis
    text = text.replace("\xa0", " ")

    # Manter quebras originais e limpar espaços à direita
    lines = [l.rstrip() for l in text.split("\n")]
    return "\n".join(lines).strip()


def list_pending_assignments() -> List[Dict[str, Any]]:
    """Retorna todas as atividades de todas as turmas que possuem correções pendentes."""
    return call_mcp("canvas_list_pending_assignments")


def list_courses() -> List[Dict[str, Any]]:
    """Lista as disciplinas do professor."""
    return call_mcp("canvas_list_courses")


def list_students(course_id: str) -> List[Dict[str, Any]]:
    """Lista oficial de matriculados da turma."""
    return call_mcp("canvas_list_students", {"course_id": str(course_id)})


def get_assignment(course_id: str, assignment_id: str) -> Dict[str, Any]:
    """Obtém detalhes completos e enunciado oficial da atividade."""
    return call_mcp("canvas_get_assignment", {
        "course_id": str(course_id),
        "assignment_id": str(assignment_id)
    })


def list_assignments(course_id: str) -> List[Dict[str, Any]]:
    """Lista todas as atividades de uma disciplina com detalhes de status e pendências."""
    return call_mcp("canvas_list_assignments", {"course_id": str(course_id)})


def get_submissions(course_id: str, assignment_id: str, only_pending: bool = True) -> List[Dict[str, Any]]:
    """Obtém submissões dos estudantes, opcionalmente filtrando apenas as pendentes de nota."""
    subs = call_mcp("canvas_get_submissions", {
        "course_id": str(course_id),
        "assignment_id": str(assignment_id),
        "only_pending": only_pending
    })
    
    # Adiciona campo 'clean_body' para facilitar análise de código
    if isinstance(subs, list):
        for s in subs:
            s["clean_body"] = clean_canvas_html(s.get("body") or "")

    return subs


def download_attachment(download_url: str, destination_path: str) -> Dict[str, Any]:
    """
    Baixa um arquivo anexado para o disco local usando a ferramenta MCP nativa.
    Garante o envio correto de 'destination_path'.
    """
    abs_path = os.path.abspath(destination_path)
    os.makedirs(os.path.dirname(abs_path), exist_ok=True)
    return call_mcp("canvas_download_attachment", {
        "download_url": download_url,
        "destination_path": abs_path
    })


def submit_grades_batch(course_id: str, assignment_id: str, grades: List[Dict[str, str]]) -> Any:
    """
    Publica notas e feedbacks em lote para múltiplos alunos.
    Formato de cada item em grades: {'user_id': '...', 'grade': '...', 'comment': '...'}
    """
    return call_mcp("canvas_submit_grades_batch", {
        "course_id": str(course_id),
        "assignment_id": str(assignment_id),
        "grades": grades
    })


def fetch_github_repo(github_url: str, destination_dir: Optional[str] = None) -> Dict[str, Any]:
    """
    Baixa e descompacta um repositório público do GitHub via ferramenta MCP nativa,
    catalogando arquivos de código e extraindo snippet do arquivo principal.
    """
    args: Dict[str, Any] = {"github_url": github_url}
    if destination_dir:
        args["destination_dir"] = os.path.abspath(destination_dir)
    return call_mcp("canvas_fetch_github_repo", args)


def prepare_workspace(course_id: str, assignment_id: str, output_dir: str = "scratch", only_pending: bool = True) -> Dict[str, Any]:
    """
    Prepara o ambiente completo para o agente:
    1. Baixa o enunciado oficial da atividade;
    2. Obtém as submissões (apenas pendentes ou todas se only_pending=False);
    3. Baixa todos os anexos automaticamente para {output_dir}/attachments/;
    4. Salva códigos submetidos via texto em {output_dir}/submissions_code/;
    5. Salva {output_dir}/prepared_submissions.json pronto para avaliação.
    """
    output_dir = os.path.abspath(output_dir)
    attachments_dir = os.path.join(output_dir, "attachments")
    code_dir = os.path.join(output_dir, "submissions_code")
    os.makedirs(attachments_dir, exist_ok=True)
    os.makedirs(code_dir, exist_ok=True)

    print(f"📦 Obtendo detalhes da atividade (Curso: {course_id}, Atividade: {assignment_id})...")
    assignment = get_assignment(course_id, assignment_id)
    with open(os.path.join(output_dir, "assignment.json"), "w", encoding="utf-8") as f:
        json.dump(assignment, f, indent=2, ensure_ascii=False)

    filter_desc = "pendentes" if only_pending else "todas (incluindo já avaliadas)"
    print(f"📥 Puxando submissões ({filter_desc})...")
    submissions = get_submissions(course_id, assignment_id, only_pending=only_pending)

    text_extensions = (".c", ".h", ".cpp", ".hpp", ".py", ".js", ".java", ".txt", ".sql", ".html", ".css", ".json")

    total_attachments = 0
    total_saved_codes = 0

    for idx, s in enumerate(submissions):
        uid = str(s.get("user_id"))
        uname = s.get("user_name", "Aluno")
        clean_name = re.sub(r"[^\w\.-]", "_", uname)

        # 1. Se aluno digitou/colou código no Canvas (online_text_entry)
        body = s.get("clean_body", "")
        if body and len(body.strip()) > 0:
            # Detectar extensão mais provável (default .c para C/ED)
            ext = ".c"
            if "def " in body or "import " in body and "#include" not in body:
                ext = ".py"
            code_filename = f"{uid}_{clean_name}{ext}"
            code_path = os.path.join(code_dir, code_filename)
            with open(code_path, "w", encoding="utf-8") as cf:
                cf.write(body)
            s["local_code_file"] = code_path
            total_saved_codes += 1

        # 2. Anexos
        attachments = s.get("attachments") or []
        for att in attachments:
            url = att.get("url")
            raw_filename = att.get("filename") or "anexo"
            clean_filename = re.sub(r"[^\w\.-]", "_", raw_filename)
            local_filename = f"{uid}_{clean_filename}"
            local_path = os.path.join(attachments_dir, local_filename)

            print(f"  [{idx+1}/{len(submissions)}] Baixando anexo de {uname}: {raw_filename}...")
            try:
                download_attachment(url, local_path)
                att["local_file"] = local_path
                total_attachments += 1

                # Se for arquivo de texto legível, carregar o conteúdo diretamente
                if local_path.lower().endswith(text_extensions) and os.path.exists(local_path):
                    try:
                        with open(local_path, "r", encoding="utf-8", errors="ignore") as af:
                            att["file_content"] = af.read()
                    except Exception as fe:
                        att["file_content_error"] = str(fe)
            except Exception as e:
                print(f"  ⚠️ Erro ao baixar anexo ({raw_filename}): {e}")
                att["download_error"] = str(e)

        # 3. Detecção e download automático de repositórios GitHub
        gh_match = re.search(r"(?:https?://)?(?:www\.)?github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)", s.get("url") or "")
        if not gh_match and s.get("clean_body"):
            gh_match = re.search(r"(?:https?://)?(?:www\.)?github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)", s["clean_body"])
        if not gh_match and s.get("body"):
            gh_match = re.search(r"(?:https?://)?(?:www\.)?github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)", s["body"])

        if gh_match:
            owner, repo = gh_match.group(1), gh_match.group(2).rstrip(".git")
            canonical_gh_url = f"https://github.com/{owner}/{repo}"
            s["github_url"] = canonical_gh_url
            gh_dest = os.path.join(output_dir, "github_repos", f"{uid}_{clean_name}")
            print(f"  [{idx+1}/{len(submissions)}] Baixando repositório GitHub de {uname}: {canonical_gh_url}...")
            try:
                repo_res = fetch_github_repo(canonical_gh_url, destination_dir=gh_dest)
                s["github_repo"] = repo_res
                s["has_code"] = True
            except Exception as e:
                print(f"  ⚠️ Erro ao baixar repositório GitHub ({canonical_gh_url}): {e}")
                s["github_error"] = str(e)

    prepared_file = os.path.join(output_dir, "prepared_submissions.json")
    with open(prepared_file, "w", encoding="utf-8") as f:
        json.dump({
            "course_id": str(course_id),
            "assignment_id": str(assignment_id),
            "assignment_name": assignment.get("name"),
            "points_possible": assignment.get("points_possible"),
            "description_html": assignment.get("description"),
            "only_pending": only_pending,
            "total_submissions": len(submissions),
            "submissions": submissions
        }, f, indent=2, ensure_ascii=False)

    total_repos = sum(1 for s in submissions if s.get("github_repo"))
    print(f"✅ Workspace preparado com sucesso!")
    print(f"   - Total de submissões carregadas: {len(submissions)}")
    print(f"   - Total de códigos de texto salvos em {code_dir}: {total_saved_codes}")
    print(f"   - Total de anexos baixados em {attachments_dir}: {total_attachments}")
    print(f"   - Total de repositórios GitHub clonados/processados: {total_repos}")
    print(f"   - Arquivo consolidado: {prepared_file}")

    return {
        "prepared_file": prepared_file,
        "total_submissions": len(submissions),
        "total_saved_codes": total_saved_codes,
        "total_attachments": total_attachments,
        "total_repos": total_repos
    }


def publish_grades_from_file(course_id: str, assignment_id: str, json_path: str) -> Any:
    """
    Valida e publica notas a partir de um arquivo JSON estruturado.
    Formato esperado: lista de objetos contendo 'user_id', 'grade' e 'comment'.
    """
    if not os.path.exists(json_path):
        raise FileNotFoundError(f"Arquivo de notas não encontrado: {json_path}")

    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    # Se for dicionário com chave 'grades', extrair
    if isinstance(data, dict) and "grades" in data:
        items = data["grades"]
    elif isinstance(data, list):
        items = data
    else:
        raise ValueError("O JSON de notas deve conter uma lista ou um objeto com a chave 'grades'.")

    validated_grades = []
    for idx, item in enumerate(items):
        uid = item.get("user_id")
        grade = item.get("grade")
        comment = item.get("comment", "")
        if not uid or grade is None:
            raise ValueError(f"Item #{idx+1} inválido: requer 'user_id' e 'grade'. Conteúdo: {item}")
        validated_grades.append({
            "user_id": str(uid),
            "grade": str(grade),
            "comment": str(comment)
        })

    print(f"🚀 Enviando lote de {len(validated_grades)} notas para o Canvas...")
    result = submit_grades_batch(course_id, assignment_id, validated_grades)
    print("✅ Lote submetido com sucesso!")
    return result


def clean_workspace(target_dir: str = "scratch") -> None:
    """
    Exclui todos os arquivos e pastas de scratch mantendo a pasta vazia.
    """
    target_path = os.path.abspath(os.path.join(PROJECT_ROOT, target_dir))
    if not os.path.exists(target_path):
        os.makedirs(target_path, exist_ok=True)
        return

    for item in os.listdir(target_path):
        item_path = os.path.join(target_path, item)
        if os.path.isfile(item_path) or os.path.islink(item_path):
            os.unlink(item_path)
        elif os.path.isdir(item_path):
            shutil.rmtree(item_path)

    print(f"🧹 Diretório temporário {target_dir}/ limpo com sucesso.")


def main():
    if len(sys.argv) < 2:
        print("Uso: python3 canvas_cli.py <comando> [argumentos...]")
        print("Comandos disponíveis:")
        print("  pending                                      - Lista todas as atividades com notas pendentes")
        print("  courses                                      - Lista as disciplinas do professor")
        print("  students <course_id>                         - Lista alunos de uma disciplina")
        print("  assignments <course_id>                      - Lista todas as tarefas de um curso com status")
        print("  assignment <course_id> <assignment_id>       - Exibe o enunciado da tarefa")
        print("  submissions <course_id> <assignment_id> [--all] - Lista submissões (pendentes ou todas)")
        print("  prepare <course_id> <assignment_id> [--all]  - Baixa tudo (enunciado, códigos e anexos) para scratch/")
        print("  publish <course_id> <assignment_id> <file>   - Publica notas em lote a partir de JSON")
        print("  github <url> [dest_dir]                      - Baixa e resume repositório GitHub para análise")
        print("  clean                                        - Limpa resíduos da pasta scratch/")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "pending":
        pendings = list_pending_assignments()
        print(json.dumps(pendings, indent=2, ensure_ascii=False))

    elif cmd == "courses":
        courses = list_courses()
        print(json.dumps(courses, indent=2, ensure_ascii=False))

    elif cmd == "students":
        if len(sys.argv) < 3:
            print("Uso: python3 canvas_cli.py students <course_id>")
            sys.exit(1)
        students = list_students(sys.argv[2])
        print(json.dumps(students, indent=2, ensure_ascii=False))

    elif cmd == "assignments":
        if len(sys.argv) < 3:
            print("Uso: python3 canvas_cli.py assignments <course_id>")
            sys.exit(1)
        assigns = list_assignments(sys.argv[2])
        print(f"\n📋 Atividades da disciplina {sys.argv[2]}:")
        for a in assigns:
            aid = a.get("id")
            aname = a.get("name")
            pts = a.get("points_possible")
            ng = a.get("needs_grading_count", 0)
            ge = a.get("graded_submissions_exist", False)
            due = a.get("due_at") or "Sem data"
            print(f"  • ID: {aid} | Pontos: {pts} | Pendentes: {ng} | Já avaliadas: {ge} | Entrega: {due}")
            print(f"    Nome: {aname}")

    elif cmd == "assignment":
        if len(sys.argv) < 4:
            print("Uso: python3 canvas_cli.py assignment <course_id> <assignment_id>")
            sys.exit(1)
        res = get_assignment(sys.argv[2], sys.argv[3])
        print(json.dumps(res, indent=2, ensure_ascii=False))

    elif cmd == "submissions":
        if len(sys.argv) < 4:
            print("Uso: python3 canvas_cli.py submissions <course_id> <assignment_id> [--all]")
            sys.exit(1)
        only_pending = "--all" not in sys.argv
        res = get_submissions(sys.argv[2], sys.argv[3], only_pending=only_pending)
        desc = "pendentes" if only_pending else "totais"
        print(f"Total de submissões {desc}: {len(res)}")
        for idx, s in enumerate(res):
            grade_str = f" | Nota: {s.get('grade')}" if s.get('grade') is not None else " | Sem nota"
            print(f"[{idx+1}] {s.get('user_name')} (ID: {s.get('user_id')}) - Tipo: {s.get('submission_type')}{grade_str}")

    elif cmd == "prepare":
        if len(sys.argv) < 4:
            print("Uso: python3 canvas_cli.py prepare <course_id> <assignment_id> [--all]")
            sys.exit(1)
        only_pending = "--all" not in sys.argv
        prepare_workspace(sys.argv[2], sys.argv[3], only_pending=only_pending)

    elif cmd == "publish":
        if len(sys.argv) < 5:
            print("Uso: python3 canvas_cli.py publish <course_id> <assignment_id> <path_to_grades.json>")
            sys.exit(1)
        res = publish_grades_from_file(sys.argv[2], sys.argv[3], sys.argv[4])
        print(json.dumps(res, indent=2, ensure_ascii=False))

    elif cmd == "github":
        if len(sys.argv) < 3:
            print("Uso: python3 canvas_cli.py github <url> [dest_dir]")
            sys.exit(1)
        dest = sys.argv[3] if len(sys.argv) > 3 else None
        res = fetch_github_repo(sys.argv[2], destination_dir=dest)
        print(json.dumps(res, indent=2, ensure_ascii=False))

    elif cmd == "clean":
        clean_workspace()

    elif cmd == "status":
        if len(sys.argv) < 3:
            print("Uso: python3 canvas_cli.py status <course_id> [assignment_id] [--json]")
            sys.exit(1)
        from check_graded import print_course_status, print_single_assignment
        as_json = "--json" in sys.argv
        clean_args = [a for a in sys.argv[2:] if not a.startswith("--")]
        course_id = clean_args[0]
        if len(clean_args) >= 2:
            print_single_assignment(course_id, clean_args[1], as_json=as_json)
    elif cmd == "modules":
        if len(sys.argv) < 3:
            print("Uso: python3 canvas_cli.py modules <course_id>")
            sys.exit(1)
        res = call_mcp("canvas_list_modules", {"course_id": str(sys.argv[2])})
        print(json.dumps(res, indent=2, ensure_ascii=False))

    else:
        print(f"Comando desconhecido: {cmd}")
        sys.exit(1)


if __name__ == "__main__":
    main()
