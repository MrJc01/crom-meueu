# Arquitetura de Dados: The Node System

## O Conceito Unificado
Para permitir que qualquer frontend consuma os dados de forma criativa, abolimos a distinção rígida entre "Posts", "Comentários" e "Artigos". Tudo é um Node.

## Estrutura do Node
Todo conteúdo na rede possui:

- **Payload**: JSON flexível (pode conter texto markdown, URL de vídeo, metadados de áudio).
- **Tipo**: Tag que define como renderizar (text/article, video/external, interaction/vote).
- **Assinatura**: Prova criptográfica de autoria.

## O Papel do Frontend
O Frontend atua como um navegador de dados. Ele não contém lógica de negócios pesada, mas contém toda a Lógica de Curadoria.

Exemplo: Um frontend pode decidir ignorar todos os Nodes do tipo `video` e exibir apenas `text`.
