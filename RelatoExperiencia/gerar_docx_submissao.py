#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Gerador de DOCX Oficial com 2 colunas nativas a partir do template do 12º Fórum Rondoniense
"""
import zipfile
import xml.etree.ElementTree as ET
import os

TEMPLATE_PATH = "Modelo_Template_Resumo Expandido e Relato Experiência-OTH.docx"
OUTPUT_PATH = "Relato_Experiencia_12_Forum_Rondoniense_OFICIAL.docx"

W_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
ET.register_namespace('w', W_NS)

def make_p(text, style_type="body", bold=False, italic=False, underline=False):
    p = ET.Element(f"{{{W_NS}}}p")
    pPr = ET.SubElement(p, f"{{{W_NS}}}pPr")
    
    spacing = ET.SubElement(pPr, f"{{{W_NS}}}spacing")
    spacing.set(f"{{{W_NS}}}line", "240")
    spacing.set(f"{{{W_NS}}}lineRule", "auto")
    
    ind = ET.SubElement(pPr, f"{{{W_NS}}}ind")
    ind.set(f"{{{W_NS}}}right", "-74")
    
    jc = ET.SubElement(pPr, f"{{{W_NS}}}jc")
    
    if style_type == "h2":
        spacing.set(f"{{{W_NS}}}before", "200")
        spacing.set(f"{{{W_NS}}}after", "60")
        jc.set(f"{{{W_NS}}}val", "left")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        b = ET.SubElement(rPr, f"{{{W_NS}}}b")
        b.set(f"{{{W_NS}}}val", "1")
        u = ET.SubElement(rPr, f"{{{W_NS}}}u")
        u.set(f"{{{W_NS}}}val", "single")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "24")
    elif style_type == "h3":
        spacing.set(f"{{{W_NS}}}before", "140")
        spacing.set(f"{{{W_NS}}}after", "40")
        jc.set(f"{{{W_NS}}}val", "both")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        b = ET.SubElement(rPr, f"{{{W_NS}}}b")
        b.set(f"{{{W_NS}}}val", "1")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "24")
    elif style_type == "ref":
        spacing.set(f"{{{W_NS}}}before", "40")
        spacing.set(f"{{{W_NS}}}after", "60")
        jc.set(f"{{{W_NS}}}val", "both")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "22")
    elif style_type == "caption":
        spacing.set(f"{{{W_NS}}}before", "100")
        spacing.set(f"{{{W_NS}}}after", "20")
        jc.set(f"{{{W_NS}}}val", "left")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        b = ET.SubElement(rPr, f"{{{W_NS}}}b")
        b.set(f"{{{W_NS}}}val", "1")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "19")
    elif style_type == "source":
        spacing.set(f"{{{W_NS}}}before", "20")
        spacing.set(f"{{{W_NS}}}after", "100")
        jc.set(f"{{{W_NS}}}val", "left")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "16")
    else:
        spacing.set(f"{{{W_NS}}}after", "80")
        ind.set(f"{{{W_NS}}}firstLine", "709")
        jc.set(f"{{{W_NS}}}val", "both")
        rPr = ET.SubElement(pPr, f"{{{W_NS}}}rPr")
        rFonts = ET.SubElement(rPr, f"{{{W_NS}}}rFonts")
        rFonts.set(f"{{{W_NS}}}ascii", "Times New Roman")
        rFonts.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
        sz = ET.SubElement(rPr, f"{{{W_NS}}}sz")
        sz.set(f"{{{W_NS}}}val", "24")

    r = ET.SubElement(p, f"{{{W_NS}}}r")
    rPr_run = ET.SubElement(r, f"{{{W_NS}}}rPr")
    rFonts_run = ET.SubElement(rPr_run, f"{{{W_NS}}}rFonts")
    rFonts_run.set(f"{{{W_NS}}}ascii", "Times New Roman")
    rFonts_run.set(f"{{{W_NS}}}hAnsi", "Times New Roman")
    
    if bold:
        b = ET.SubElement(rPr_run, f"{{{W_NS}}}b")
        b.set(f"{{{W_NS}}}val", "1")
    if italic:
        i = ET.SubElement(rPr_run, f"{{{W_NS}}}i")
        i.set(f"{{{W_NS}}}val", "1")
    if underline:
        u = ET.SubElement(rPr_run, f"{{{W_NS}}}u")
        u.set(f"{{{W_NS}}}val", "single")
        
    t = ET.SubElement(r, f"{{{W_NS}}}t")
    t.set("{http://www.w3.org/XML/1998/namespace}space", "preserve")
    t.text = text
    
    return p

def set_p_text(p, text):
    t_nodes = p.findall(f".//{{{W_NS}}}t")
    if t_nodes:
        t_nodes[0].text = text
        for extra in t_nodes[1:]:
            extra.text = ""
    else:
        r = ET.SubElement(p, f"{{{W_NS}}}r")
        t = ET.SubElement(r, f"{{{W_NS}}}t")
        t.text = text

def build_docx():
    import html.parser
    
    class ArticleExtractor(html.parser.HTMLParser):
        def __init__(self):
            super().__init__()
            self.in_main = False
            self.current_tag = None
            self.current_attrs = {}
            self.current_text = []
            self.elements = []
            
        def handle_starttag(self, tag, attrs):
            attrs_dict = dict(attrs)
            if tag == 'main' and 'content-two-cols' in attrs_dict.get('class', ''):
                self.in_main = True
                return
            if not self.in_main:
                return
            if tag in ['h2', 'h3', 'p', 'div']:
                self.current_tag = tag
                self.current_attrs = attrs_dict
                self.current_text = []
                
        def handle_endtag(self, tag):
            if tag == 'main':
                self.in_main = False
                return
            if not self.in_main:
                return
            if tag in ['h2', 'h3', 'p', 'div'] and self.current_text:
                text = "".join(self.current_text).strip()
                if text:
                    self.elements.append((tag, self.current_attrs.get('class', ''), text))
                self.current_text = []
                self.current_tag = None
                
        def handle_data(self, data):
            if self.in_main and self.current_tag:
                self.current_text.append(data)

    with open("relato_experiencia_forum_rondoniense.html", "r", encoding="utf-8") as f:
        html_content = f.read()

    parser = ArticleExtractor()
    parser.feed(html_content)

    with zipfile.ZipFile(TEMPLATE_PATH, 'r') as zin:
        doc_xml = zin.read('word/document.xml')
        root = ET.fromstring(doc_xml)
        ns = {'w': W_NS}
        body = root.find('w:body', ns)
        children = list(body)
        
        # 1. Update Header Information in children[7, 8, 10, 12, 14, 15, 16, 17]
        set_p_text(children[7], "Eixo Temático: 3. Inteligência Artificial, Transformação Digital e Tecnologias Emergentes")
        set_p_text(children[8], "Linha Temática: II. Sistemas Inteligentes, Automação de Processos e Tomada de Decisão")
        set_p_text(children[10], "Mediação inteligente na gestão docente universitária: relato de experiência da integração de agente autônomo ao Canvas LMS via Model Context Protocol (MCP) com interface via Telegram")
        set_p_text(children[12], "Karan Luciano1*, Equipe Docente de Tecnologias Educacionais1, Coordenação de Pesquisa e Inovação Acadêmica1")
        set_p_text(children[14], "1Curso de Bacharelado em Ciência da Computação, Afya Centro Universitário de Ji-Paraná, Ji-Paraná, Rondônia, Brasil")
        set_p_text(children[15], "")
        set_p_text(children[16], "")
        set_p_text(children[17], "*Autor(a) correspondente: Endereço profissional: Av. Universitária, s/n, Aurélio Bernardi, Ji-Paraná - RO, CEP: 76907-438. E-mail: karan.luciano@afya.com.br")
        
        header_elements = children[:20]
        final_sectPr = children[-1]
        
        body.clear()
        for h in header_elements:
            body.append(h)
            
        for tag, cls, text in parser.elements:
            if tag == 'h2':
                body.append(make_p(text, style_type="h2", bold=True, underline=True))
            elif tag == 'h3':
                body.append(make_p(text, style_type="h3", bold=True))
            elif cls == 'figure-caption' or cls == 'table-title':
                body.append(make_p(text, style_type="caption", bold=True))
            elif cls == 'figure-source' or cls == 'table-footer':
                body.append(make_p(text, style_type="source"))
            elif cls == 'reference-entry':
                body.append(make_p(text, style_type="ref"))
            elif tag == 'p':
                body.append(make_p(text, style_type="body"))
                
        body.append(final_sectPr)

        new_doc_xml = ET.tostring(root, encoding='utf-8', xml_declaration=True)

    with zipfile.ZipFile(TEMPLATE_PATH, 'r') as zin, zipfile.ZipFile(OUTPUT_PATH, 'w') as zout:
        for item in zin.infolist():
            if item.filename == 'word/document.xml':
                zout.writestr(item, new_doc_xml)
            else:
                zout.writestr(item, zin.read(item.filename))

    print(f"Sucesso! Documento gerado com 2 colunas nativas em: {OUTPUT_PATH}")

if __name__ == '__main__':
    build_docx()
