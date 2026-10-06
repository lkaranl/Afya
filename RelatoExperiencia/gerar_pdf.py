#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script utilitário para gerar o PDF do artigo com o banner overlay integrado na página 1
e o rodapé oficial da ProPPEXI aplicado em todas as páginas.
"""
import os
import subprocess
import pypdf

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
HTML_PATH = os.path.join(BASE_DIR, "relato_experiencia_forum_rondoniense.html")
BANNER_PDF = os.path.join(BASE_DIR, "banner_overlay.pdf")
FOOTER_PDF = os.path.join(BASE_DIR, "footer_overlay.pdf")
TEMP_PDF = os.path.join(BASE_DIR, "_temp_artigo.pdf")
OUTPUT_PDF = os.path.join(BASE_DIR, "artigo_forum.pdf")
CHROME_BIN = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

def main():
    print("1. Compilando HTML via Google Chrome headless...")
    cmd = [
        CHROME_BIN,
        "--headless",
        "--disable-gpu",
        f"--print-to-pdf={TEMP_PDF}",
        "--no-pdf-header-footer",
        f"file://{HTML_PATH}"
    ]
    subprocess.run(cmd, check=True)

    print("2. Aplicando banner (pág 1) e rodapé oficial em TODAS as páginas...")
    article = pypdf.PdfReader(TEMP_PDF)
    banner = pypdf.PdfReader(BANNER_PDF)
    footer = pypdf.PdfReader(FOOTER_PDF)
    writer = pypdf.PdfWriter()

    for idx, p in enumerate(article.pages):
        # Aplicar banner apenas na primeira página
        if idx == 0:
            p.merge_page(banner.pages[0], over=True)
        # Aplicar rodapé oficial em TODAS as páginas
        p.merge_page(footer.pages[0], over=True)
        writer.add_page(p)

    with open(OUTPUT_PDF, "wb") as f:
        writer.write(f)

    if os.path.exists(TEMP_PDF):
        os.remove(TEMP_PDF)

    final = pypdf.PdfReader(OUTPUT_PDF)
    print(f"Sucesso! PDF gerado em: {OUTPUT_PDF}")
    print(f"Total de páginas: {len(final.pages)}")

if __name__ == "__main__":
    main()
