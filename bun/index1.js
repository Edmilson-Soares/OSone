import { MQ }  from './mq-client'; // Assumindo que o código está no arquivo mq-client.js

async function main() {
    // 1. Criar uma instância do cliente MQ
    const client = new MQ();

    try {
        // 2. Conectar ao servidor MQ
        await client.connect('mq://root:admin@localhost:4222');
        console.log('Conectado com sucesso! ID:', client.ID);

        // 3. Assinar um tópico
        await client.subscribe('notificacoes', (message, topic) => {
            console.log(`\nNova mensagem no tópico ${topic}:`, message);
        });

        // 4. Publicar uma mensagem
        await client.publish('notificacoes', 'Olá do Node.js!');
        console.log('Mensagem publicada com sucesso!');

        // 5. Criar um serviço
        await client.service('calculadora', (data, reply) => {
            console.log('\nSolicitação recebida no serviço calculadora:', data);
            try {
                const { a, b, op } = JSON.parse(data);
                let result;
                
                switch (op) {
                    case '+': result = a + b; break;
                    case '-': result = a - b; break;
                    case '*': result = a * b; break;
                    case '/': result = a / b; break;
                    default: throw new Error('Operação inválida');
                }
                
                reply(null, JSON.stringify({ result }));
            } catch (err) {
                reply(err.message, null);
            }
        });

        // 6. Fazer uma requisição para um serviço
        const response = await client.request('calculadora', JSON.stringify({
            a: 10,
            b: 5,
            op: '+'
        }), 3000);
        
        console.log('\nResposta do serviço calculadora:', JSON.parse(response));

        // Manter a conexão aberta por um tempo
        console.log('\nAguardando mensagens... (Ctrl+C para sair)');
        await new Promise(resolve => setTimeout(resolve, 60000));

    } catch (err) {
        console.error('Erro:', err.message);
    } finally {
        // 7. Desconectar
        client.disconnect();
        console.log('Conexão encerrada.');
    }
}

// Executar o exemplo
main().catch(console.error);