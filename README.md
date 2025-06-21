# Batedor — Monitoramento Profissional de Sistemas em Go
<p align="left">
    <img src="https://img.shields.io/badge/versão-v1.0-blue.svg" alt="Versão">
    <img src="https://img.shields.io/badge/licença-GLP3-blue.svg" alt="Licença">
    <img src="https://img.shields.io/badge/Go-1.18%2B-cyan.svg" alt="Go Version">
    <img src="https://img.shields.io/badge/plataforma-Linux-blue.svg" alt="Plataforma">
    <img src="https://img.shields.io/badge/feito_no-Brasil-blue.svg" alt="Feito no Brasil">
</p>

## 🚀 O que é o Batedor?

O **Batedor** é uma ferramenta profissional desenvolvida em **Go** para monitoramento e diagnóstico em tempo real de sistemas Linux. Com interface TUI (terminal interativa), dashboard Web e histórico persistente de métricas, ele fornece uma visão completa e robusta dos recursos do seu servidor, ideal para sysadmins e devops.

### 🎬 Funcionamento do Bot

Veja abaixo uma demonstração visual do funcionamento do Batedor:

![Funcionamento do Bot](https://github.com/henriquetourinho/batedor/blob/main/media/funcionamento.gif?raw=true)

---

## 🛠️ Instalação e Uso

Siga os passos para ter o Batedor operacional em seu sistema.

### 1. Pré-requisitos

Garanta que as seguintes ferramentas estejam instaladas no seu sistema (Debian/Ubuntu):

```bash
sudo apt update && sudo apt install git golang build-essential dmidecode
```

### 2. Instalação

Clone o repositório e deixe o Go cuidar das dependências.

```bash
# Clone o projeto
git clone https://github.com/henriquetourinho/batedor.git

# Entre na pasta do projeto
cd batedor

# Baixe e organize todas as dependências do Go
go mod tidy
```

### 3. Execução

O Batedor possui dois modos de operação (apenas com `go run`):

#### Modo Padrão (Apenas Terminal):

```bash
go run .
```

#### Modo Híbrido (Terminal + Web):

```bash
go run . --web
```

E então acesse [http://localhost:9090](http://localhost:9090) no seu navegador.

---

## ⌨️ Comandos e Atalhos

| Tecla | Tela Principal                | Tela de Histórico (H)        |
|-------|------------------------------|------------------------------|
| Q     | Sair do programa             | Voltar para a tela principal |
| C     | Ordenar processos por CPU    | Alternar para o gráfico de CPU|
| M     | Ordenar processos por Memória| Alternar para o gráfico de Memória|
| P     | Ordenar processos por PID    | -                            |
| K     | Encerrar ("Kill") o processo selecionado | -                  |
| H     | Abrir tela de Histórico      | -                            |
| F1    | Abrir a tela de Ajuda        | -                            |

Na tela de Ajuda, qualquer tecla pressionada te levará de volta à tela principal.

---

## 🧩 Recursos Profissionais

- **Monitoramento em tempo real:** CPU (núcleo a núcleo), memória, disco, rede, processos, informações do host.
- **Interface TUI amigável:** gráficos, tabelas, histórico, atalhos.
- **Dashboard Web:** visualização instantânea e responsiva via navegador.
- **Histórico persistente:** métricas armazenadas em SQLite local.
- **Gestão de processos:** filtro, ordenação, kill seguro com confirmação.
- **Visualização de rede:** IP público, latência, interface principal, tráfego.
- **Ajuda integrada:** manual de comandos e atalhos acessível por F1.
- **Execução multiplataforma** (Linux).
- **Código limpo, modular e fácil de estender**.

---

## 🔐 Segurança e Boas Práticas

- Recomenda-se execução como root para acesso total aos dados do sistema.
- Nenhuma coleta ou envio externo de informações.
- Encerramento de processos com confirmação.
- Banco de dados local, sem sobrescrita de dados sem confirmação.

---

## 🤝 Apoie o Projeto

Se o **Batedor** te ajudou, considere apoiar para manter a iniciativa ativa e evoluindo para toda a comunidade:

**Chave Pix:**  
```
poupanca@henriquetourinho.com.br
```

---

### Licença

Este projeto é distribuído sob a **GPL-3.0 license**. Veja o arquivo `LICENSE` para mais detalhes.

## 🙋‍♂️ Desenvolvido por

**Carlos Henrique Tourinho Santana** 📍 Salvador - Bahia  
<br>
🔗 Wiki Debian: [wiki.debian.org/henriquetourinho](https://wiki.debian.org/henriquetourinho)  
<br>
🔗 LinkedIn: [br.linkedin.com/in/carloshenriquetourinhosantana](https://br.linkedin.com/in/carloshenriquetourinhosantana)  
<br>
🔗 GitHub: [github.com/henriquetourinho](https://github.com/henriquetourinho). fale para o webdesign, o que ele tem que fazer num site de apresentação e tal.; tiudo que ele debe coplocar. a tudo deve ficar em index.html. 
