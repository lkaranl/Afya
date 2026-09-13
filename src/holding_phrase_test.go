package main

import "testing"

func TestIsUnfinishedExecution(t *testing.T) {
	unfinishedExamples := []string{
		"Um momento, por favor.",
		"um momento por favor",
		"Um momento...",
		"Aguarde um momento",
		"Aguarde um instante!",
		"Aguarde, por favor.",
		"Estou verificando no Canvas...",
		"Só um instante.",
		"Processando sua solicitação...",
		// Caso real relatado pelo usuário:
		"Aguarde um momento enquanto localizo os módulos nas suas turmas ativas. Assim que identificar os módulos corretos, criarei a página com o conteúdo sobre linguagens compiladas e interpretadas e a vincularei.\n\nObservação: Para garantir que a página seja inserida no módulo \"Módulo Introdutório\" de ambas as turmas, preciso primeiro verificar se este módulo já existe. Caso não exista em alguma das turmas, criarei o módulo e depois inserirei a página.\n\nAguarde enquanto realizo as verificações e ações necessárias.",
	}

	for _, ex := range unfinishedExamples {
		if !isUnfinishedExecution(ex) {
			t.Errorf("Esperava que %q fosse detectado como execução inacabada / frase de espera, mas não foi", ex)
		}
	}

	validResponses := []string{
		"Olá, professor! Como posso ajudar hoje?",
		"Concluí a criação da página 'Linguagens Compiladas e Interpretadas' e já vinculei ao módulo nas suas duas turmas ativas com sucesso!",
		"| Aluno | ID | Nota |\n| :--- | :--- | :--- |\n| João | 123 | 90 |",
		"Nenhuma pendência de correção foi encontrada.",
		"Professor, elaborei a proposta da atividade sobre Recursão em C com 20 pontos no módulo 'Semana 4'. Segue o enunciado:\n\n[Enunciado]\n\nDeseja que eu publique agora no Canvas LMS ou gostaria de fazer algum ajuste?",
		"Professor, preparei o simulado ENADE com 5 questões. Posso publicar na turma de Estrutura de Dados?",
	}

	for _, vr := range validResponses {
		if isUnfinishedExecution(vr) {
			t.Errorf("Não esperava que %q fosse detectado como execução inacabada", vr)
		}
	}
}
