#!/usr/bin/env python3
"""
Gerador de Tabela de Revisão para o Professor
Lê um arquivo JSON de notas (ex: scratch/grades_draft.json) e imprime a tabela
Markdown perfeitamente formatada para o ponto de parada de aprovação humana.
Também permite exportar diretamente o JSON validado pronto para publicação no Canvas.
"""

import sys
import os
import json


def generate_markdown_table(json_path: str, export_publish_path: str = None) -> None:
    if not os.path.exists(json_path):
        print(f"Erro: Arquivo {json_path} não encontrado.")
        sys.exit(1)

    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    if isinstance(data, dict) and "grades" in data:
        items = data["grades"]
    elif isinstance(data, list):
        items = data
    else:
        print("Erro: JSON inválido. Esperada lista de notas ou objeto com chave 'grades'.")
        sys.exit(1)

    # Separar os que têm submissão e os sem submissão se o campo existir
    submitted = [it for it in items if it.get("has_submission", True)]
    unsubmitted = [it for it in items if not it.get("has_submission", True)]

    print(f"\n### 📊 Tabela de Avaliação Sugerida ({len(submitted)} Alunos com Envio)\n")
    print("| Aluno | ID | Nota Sugerida / Máxima | Resumo do Feedback |")
    print("| :--- | :---: | :---: | :--- |")

    grades_numeric = []
    perfect_count = 0
    errors_count = 0

    for item in submitted:
        name = item.get("student_name") or item.get("user_name") or f"Aluno {item.get('user_id')}"
        uid = item.get("user_id", "-")
        grade = str(item.get("grade", "0"))
        max_grade = str(item.get("max_grade", "100"))
        summary = item.get("summary") or item.get("comment", "")[:120]

        # Sanitizar pipes para não quebrar a tabela Markdown
        summary_clean = summary.replace("|", "-").replace("\n", " ").strip()

        print(f"| **{name}** | {uid} | {grade} / {max_grade} | {summary_clean} |")

        try:
            val = float(grade)
            grades_numeric.append(val)
            if val >= 100:
                perfect_count += 1
            elif val < 70:
                errors_count += 1
        except ValueError:
            pass

    if unsubmitted:
        print(f"\n> **Alunos sem envio ({len(unsubmitted)} alunos):** Mantidos com nota 0/100 e aviso de ausência de entrega no prazo.\n")

    print("---\n")
    if grades_numeric:
        avg = sum(grades_numeric) / len(grades_numeric)
        print(f"**📈 Estatísticas das Entregas:**")
        print(f"- **Total de Avaliados:** {len(grades_numeric)}")
        print(f"- **Média:** {avg:.1f} / 100")
        print(f"- **Notas Máximas (100):** {perfect_count}")
        print(f"- **Abaixo de 70:** {errors_count}\n")

    print("Professor, deseja que eu ajuste alguma nota ou posso publicar as avaliações no Canvas?")

    if export_publish_path:
        publish_payload = []
        for it in items:
            publish_payload.append({
                "user_id": str(it.get("user_id")),
                "grade": str(it.get("grade", "0")),
                "comment": it.get("comment", it.get("summary", ""))
            })
        os.makedirs(os.path.dirname(os.path.abspath(export_publish_path)), exist_ok=True)
        with open(export_publish_path, "w", encoding="utf-8") as pf:
            json.dump(publish_payload, pf, indent=2, ensure_ascii=False)
        print(f"\n📦 Payload pronto para publicação exportado para: {export_publish_path}")


def main():
    if len(sys.argv) < 2:
        print("Uso: python3 generate_review_table.py <grades.json> [--export-publish payload_pronto.json]")
        sys.exit(1)

    json_path = sys.argv[1]
    export_path = None
    if "--export-publish" in sys.argv:
        idx = sys.argv.index("--export-publish")
        if idx + 1 < len(sys.argv):
            export_path = sys.argv[idx + 1]

    generate_markdown_table(json_path, export_publish_path=export_path)


if __name__ == "__main__":
    main()
