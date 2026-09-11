#!/usr/bin/env python3
"""
Testador Automatizado de Submissões em C (Afya Canvas Assistant)
Avalia códigos de estudantes em C com rigor técnico e didática para iniciantes.

Capacidades:
1. Higienização de Markdown e comentários de linha única em submissões brutas.
2. Detecção automática e isolamento de funções (com ou sem main do aluno).
3. Análise estática de armadilhas clássicas (ex: *p++ vs (*p)++).
4. Compilação inteligente (com harness de teste ou sintaxe via objeto -c).
5. Execução protegida por timeout.
6. Geração de relatórios com apontamento cirúrgico de erros e sugestões didáticas.
"""

import sys
import os
import subprocess
import re
import json
from typing import Dict, Any, Optional, List


def check_compiler() -> str:
    """Verifica se gcc ou clang está disponível no sistema."""
    for comp in ["gcc", "clang"]:
        try:
            subprocess.run([comp, "--version"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)
            return comp
        except (subprocess.SubprocessError, FileNotFoundError):
            continue
    raise RuntimeError("Nenhum compilador C (gcc/clang) foi encontrado no sistema.")


def sanitize_c_code(raw_code: str) -> str:
    """
    Remove envoltórios de Markdown (```c ... ```) e textos explicativos antes/depois/entre o código.
    Extrai definições válidas de funções C e diretivas #include.
    Protege contra comentários // em códigos colados em uma única linha no Canvas.
    """
    if not raw_code:
        return ""

    text = raw_code

    # Se todo o código estiver em 1 linha contendo //, converter // em /* ... */
    if text.count("\n") <= 1 and "//" in text:
        text = re.sub(r'//([^;{}]+)', r'/* \1 */', text)

    # Remover blocos e tags Markdown
    text = re.sub(r'```(?:c|cpp)?', '', text)

    # Extrair diretivas #include e #define
    directives = re.findall(r'^\s*#(?:include|define)[^\n]+', text, flags=re.MULTILINE)

    # Extrair funções C válidas procurando assinaturas com { ... }
    # Padrão: tipo nome(argumentos) { corpo }
    fn_pattern = re.compile(
        r'(?:(?:static|inline|const)\s+)?(?:int|void|float|double|char\*?|bool|size_t)\s+[a-zA-Z0-9_]+\s*\([^)]*\)\s*\{',
        re.MULTILINE
    )

    extracted_blocks = []
    for m in fn_pattern.finditer(text):
        start_pos = m.start()
        open_brackets = 0
        end_pos = start_pos
        started = False

        for i in range(m.end() - 1, len(text)):
            char = text[i]
            if char == '{':
                open_brackets += 1
                started = True
            elif char == '}':
                open_brackets -= 1
                if started and open_brackets == 0:
                    end_pos = i + 1
                    break

        if end_pos > start_pos and open_brackets == 0:
            extracted_blocks.append(text[start_pos:end_pos].strip())

    if extracted_blocks:
        directives_part = "\n".join(directives) + "\n\n" if directives else ""
        return (directives_part + "\n\n".join(extracted_blocks)).strip()

    # Fallback caso não consiga casar com o parser de blocos
    lines = text.splitlines()
    code_lines = []
    for line in lines:
        sline = line.strip()
        if any(sline.startswith(p) for p in ["#include", "#define", "int ", "void ", "char ", "float ", "double ", "while", "for", "if", "return", "{", "}", "//", "/*"]):
            code_lines.append(line)
    return "\n".join(code_lines).strip()


def inspect_pointer_precedence(code: str) -> List[Dict[str, Any]]:
    """
    Detecta o bug clássico de C: *ptr++
    Em C, ++ tem maior precedência que *. Portanto, *ptr++ incrementa o ponteiro (o endereço),
    não o valor apontado. O correto é (*ptr)++ ou *ptr += 1.
    """
    issues = []
    pattern = r'(\*\s*([a-zA-Z0-9_]+)\s*\+\+)'
    for m in re.finditer(pattern, code):
        start = m.start()
        if start > 0 and code[start - 1] == '(':
            continue  # (*ptr)++ está correto
        raw_expr = m.group(1)
        var_name = m.group(2)
        issues.append({
            "type": "pointer_precedence",
            "expression": raw_expr,
            "variable": var_name,
            "message": (
                f"Uso de `{raw_expr};` em vez de `(*{var_name})++;` ou `*{var_name} += 1;`. "
                f"Em C, o operador `++` tem precedência sobre `*`, deslocando o endereço de memória "
                f"do ponteiro ao invés de incrementar a variável apontada."
            )
        })
    return issues


def evaluate_c_code(
    code_or_file: str,
    expected_function: Optional[str] = None,
    test_harness_src: Optional[str] = None,
    timeout_sec: int = 5,
    temp_dir: str = "scratch/c_eval_temp"
) -> Dict[str, Any]:
    """
    Avalia completamente um código C (passado como caminho de arquivo ou string de código).
    """
    compiler = check_compiler()
    os.makedirs(temp_dir, exist_ok=True)

    if os.path.isfile(code_or_file):
        with open(code_or_file, "r", encoding="utf-8", errors="ignore") as f:
            raw_text = f.read()
        base_name = os.path.splitext(os.path.basename(code_or_file))[0]
    else:
        raw_text = code_or_file
        base_name = "snippet"

    clean_code = sanitize_c_code(raw_text)
    ptr_issues = inspect_pointer_precedence(clean_code)

    # Detectar se o aluno incluiu main()
    has_main = bool(re.search(r'\b(?:int|void)\s+main\s*\(', clean_code))

    # Detectar nome da função implementada
    fn_name = ""
    fn_match = re.search(r'(?:int|void|float|double|char\*?|bool)\s+([a-zA-Z0-9_]+)\s*\(', clean_code)
    if fn_match:
        fn_name = fn_match.group(1)

    # Preparar código para compilação
    modified_code = clean_code
    if has_main and test_harness_src:
        # Renomear main do aluno para evitar conflito com o harness de teste
        modified_code = re.sub(r'\bint\s+main\s*\(', 'int student_main(', modified_code)
        modified_code = re.sub(r'\bvoid\s+main\s*\(', 'void student_main(', modified_code)

    # Headers essenciais se ausentes
    headers = []
    if "<stdio.h>" not in modified_code:
        headers.append("#include <stdio.h>")
    if "<stdlib.h>" not in modified_code:
        headers.append("#include <stdlib.h>")
    if "<stddef.h>" not in modified_code:
        headers.append("#include <stddef.h>")
    header_block = "\n".join(headers) + "\n" if headers else ""

    alias_block = ""
    if expected_function and fn_name and fn_name != expected_function and fn_name != "student_main":
        alias_block = f"\n#ifndef {expected_function}\n#define {expected_function} {fn_name}\n#endif\n"

    src_path = os.path.join(temp_dir, f"{base_name}_eval.c")
    bin_path = os.path.join(temp_dir, f"{base_name}_eval.out")

    # Decidir o que compilar
    if test_harness_src:
        full_src = f"{header_block}{modified_code}\n{alias_block}\n{test_harness_src}\n"
        with open(src_path, "w", encoding="utf-8") as sf:
            sf.write(full_src)
        compile_cmd = [compiler, "-Wall", src_path, "-o", bin_path]
    elif not has_main:
        # Se não tem harness e nem main, compilar apenas objeto (-c) para checar erros de sintaxe
        full_src = f"{header_block}{modified_code}\n"
        with open(src_path, "w", encoding="utf-8") as sf:
            sf.write(full_src)
        obj_path = os.path.join(temp_dir, f"{base_name}_eval.o")
        compile_cmd = [compiler, "-Wall", "-c", src_path, "-o", obj_path]
        bin_path = None
    else:
        full_src = f"{header_block}{modified_code}\n"
        with open(src_path, "w", encoding="utf-8") as sf:
            sf.write(full_src)
        compile_cmd = [compiler, "-Wall", src_path, "-o", bin_path]

    p_comp = subprocess.run(compile_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    compiles = (p_comp.returncode == 0)

    result = {
        "identifier": base_name,
        "compiles": compiles,
        "has_main": has_main,
        "detected_function": fn_name,
        "expected_function": expected_function,
        "function_name_matches": (fn_name == expected_function) if expected_function else True,
        "pointer_issues": ptr_issues,
        "compile_error": None,
        "test_output": None,
        "test_status": "not_run",
        "clean_code": clean_code
    }

    if not compiles:
        err_lines = p_comp.stderr.strip().splitlines()
        first_err = err_lines[0] if err_lines else "Erro de compilação desconhecido"
        # Extrair linha
        line_m = re.search(r":(\d+):(?:\d+:)?\s+error:\s+(.+)", p_comp.stderr)
        line_no = int(line_m.group(1)) if line_m else None
        msg = line_m.group(2) if line_m else first_err
        result["compile_error"] = {
            "line": line_no,
            "message": msg,
            "raw_stderr": p_comp.stderr
        }
        return result

    # Se compilou e temos binário executável para rodar testes
    if bin_path and os.path.exists(bin_path):
        try:
            p_run = subprocess.run([bin_path], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=timeout_sec)
            result["test_status"] = "success" if p_run.returncode == 0 else "runtime_error"
            result["test_output"] = p_run.stdout.strip()
            result["runtime_stderr"] = p_run.stderr.strip()
            result["exit_code"] = p_run.returncode
        except subprocess.TimeoutExpired:
            result["test_status"] = "timeout"
            result["test_output"] = f"Execução excedeu o limite de {timeout_sec}s (loop infinito)."
        finally:
            if os.path.exists(bin_path):
                try:
                    os.unlink(bin_path)
                except OSError:
                    pass

    return result


def evaluate_batch_directory(
    code_dir: str,
    expected_function: Optional[str] = None,
    test_harness_path: Optional[str] = None,
    output_report_path: str = "scratch/c_batch_report.json"
) -> List[Dict[str, Any]]:
    """
    Avalia todos os arquivos de código C de um diretório em lote.
    """
    if not os.path.exists(code_dir):
        print(f"Diretório não encontrado: {code_dir}")
        return []

    harness_src = None
    if test_harness_path and os.path.exists(test_harness_path):
        with open(test_harness_path, "r", encoding="utf-8") as hf:
            harness_src = hf.read()

    files = sorted([os.path.join(code_dir, f) for f in os.listdir(code_dir) if f.endswith(".c")])
    print(f"🔬 Avaliando em lote {len(files)} arquivos C em '{code_dir}'...")

    reports = []
    for idx, fpath in enumerate(files):
        fname = os.path.basename(fpath)
        print(f"  [{idx+1}/{len(files)}] Analisando {fname}...")
        res = evaluate_c_code(fpath, expected_function=expected_function, test_harness_src=harness_src)
        reports.append(res)

    os.makedirs(os.path.dirname(os.path.abspath(output_report_path)), exist_ok=True)
    with open(output_report_path, "w", encoding="utf-8") as rf:
        json.dump(reports, rf, indent=2, ensure_ascii=False)

    print(f"\n✅ Avaliação concluída! Relatório salvo em: {output_report_path}")
    return reports


def main():
    if len(sys.argv) < 2:
        print("Uso: python3 test_c_submissions.py <arquivo.c | pasta_codigos> [funcao_esperada] [harness.c]")
        print("Exemplos:")
        print("  python3 test_c_submissions.py scratch/submissions_code/ busca_produto_telemetria harness.c")
        print("  python3 test_c_submissions.py scratch/codigo_aluno.c busca_produto_telemetria")
        sys.exit(1)

    target = sys.argv[1]
    expected_fn = sys.argv[2] if len(sys.argv) > 2 else None
    harness_file = sys.argv[3] if len(sys.argv) > 3 else None

    if os.path.isdir(target):
        evaluate_batch_directory(target, expected_function=expected_fn, test_harness_path=harness_file)
    elif os.path.isfile(target):
        harness_src = None
        if harness_file and os.path.exists(harness_file):
            with open(harness_file, "r", encoding="utf-8") as hf:
                harness_src = hf.read()
        res = evaluate_c_code(target, expected_function=expected_fn, test_harness_src=harness_src)
        print(json.dumps(res, indent=2, ensure_ascii=False))
    else:
        print(f"Alvo não encontrado: {target}")
        sys.exit(1)


if __name__ == "__main__":
    main()
