#!/usr/bin/env python3
"""
Descompactador e Mapeador de ZIPs do Canvas LMS
Descompacta pacotes exportados pelo Canvas e mapeia arquivos diretamente aos
alunos e seus IDs oficiais da turma.
"""

import sys
import os
import zipfile
import re
import json
from typing import Optional, Dict

# Importa utilitário MCP se disponível
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
try:
    from canvas_cli import list_students
except ImportError:
    list_students = None


def parse_canvas_filename(filename: str) -> Dict[str, str]:
    """
    O Canvas exporta arquivos com a convenção:
    [nome_sobrenome]_[user_id]_[submission_id]_[nome_original.ext]
    ou variações com números no início.
    """
    basename = os.path.basename(filename)
    match = re.search(r"^(.*?)_(\d+)_(\d+)_(.+)$", basename)
    if match:
        return {
            "name_hint": match.group(1).replace("_", " ").title(),
            "user_id": match.group(2),
            "submission_id": match.group(3),
            "original_filename": match.group(4)
        }
    return {
        "name_hint": basename,
        "user_id": "",
        "submission_id": "",
        "original_filename": basename
    }


def unpack_and_map_zip(zip_path: str, course_id: Optional[str] = None, output_dir: str = "scratch/zip_unpacked") -> str:
    if not os.path.exists(zip_path):
        raise FileNotFoundError(f"Arquivo ZIP não encontrado: {zip_path}")

    os.makedirs(output_dir, exist_ok=True)

    print(f"📦 Descompactando {zip_path} em {output_dir}...")
    with zipfile.ZipFile(zip_path, "r") as z:
        z.extractall(output_dir)

    # Obter lista oficial de alunos se course_id foi passado e temos MCP
    students_map = {}
    if course_id and list_students:
        try:
            print(f"📋 Consultando lista oficial de alunos da turma {course_id}...")
            stud_list = list_students(course_id)
            for st in stud_list:
                uid = str(st.get("id"))
                name = st.get("name") or st.get("sortable_name")
                students_map[uid] = name
        except Exception as e:
            print(f"⚠️ Não foi possível consultar a lista de alunos via MCP: {e}")

    extracted_files = []
    for root, _, files in os.walk(output_dir):
        for f in files:
            if f.startswith(".") or f == "manifest.json":
                continue
            full_path = os.path.join(root, f)
            meta = parse_canvas_filename(f)
            uid = meta["user_id"]
            official_name = students_map.get(uid, meta["name_hint"])

            extracted_files.append({
                "user_id": uid,
                "student_name": official_name,
                "original_filename": meta["original_filename"],
                "file_path": os.path.abspath(full_path),
                "extension": os.path.splitext(f)[1].lower()
            })

    manifest_path = os.path.join(output_dir, "manifest.json")
    with open(manifest_path, "w", encoding="utf-8") as mf:
        json.dump({
            "zip_source": os.path.abspath(zip_path),
            "course_id": course_id,
            "total_files": len(extracted_files),
            "submissions": extracted_files
        }, mf, indent=2, ensure_ascii=False)

    print(f"✅ Concluído! {len(extracted_files)} arquivos processados.")
    print(f"📄 Manifesto gerado em: {manifest_path}")
    return manifest_path


def main():
    if len(sys.argv) < 2:
        print("Uso: python3 unpack_canvas_zip.py <arquivo.zip> [course_id] [diretorio_destino]")
        sys.exit(1)

    zip_file = sys.argv[1]
    course_id = sys.argv[2] if len(sys.argv) > 2 else None
    out_dir = sys.argv[3] if len(sys.argv) > 3 else "scratch/zip_unpacked"

    unpack_and_map_zip(zip_file, course_id, out_dir)


if __name__ == "__main__":
    main()
