import { MQ } from './mq-client'; // Assumindo que o código está no arquivo mq-client.js

async function main() {
    // 1. Criar uma instância do cliente MQ
    const client = new MQ();

    try {
        // 2. Conectar ao servidor MQ

      //  await client.connect('mq://apikey:8rQJ0Afl1sY3NfC57b67X4sgrciuUoVWFtKO4IZQSH_4HsfkVb8yYJAXfCF_nyJTdIpG0C1_IcCfG_Jn_5fJzj2W5osg3A90nPnq3cMi2Oysj8SNU_yuNSDfVw==@localhost:4222');
        
     //   await client.connect('mq://root:chave-secreta-32-bytes-123456789@37.27.39.202:4052');
      
      //console.log('Conectado com sucesso! ID:', client.ID);

        //  await client.connect('mq://apikey:8rQJ0Afl1sY3NfC57b67X4sgrciuUoVWFtKO4IZQSH_4HsfkVb8yYJAXfCF_nyJTdIpG0C1_IcCfG_Jn_5fJzj2W5osg3A90nPnq3cMi2Oysj8SNU_yuNSDfVw==@localhost:4222');

        await client.connect('mq://root:chave-secreta-32-bytes-123456789@37.27.39.202:4052');

        console.log('Conectado com sucesso! ID:', client.ID);

        // 3. Assinar um tópico
        await client.subscribe('testtttrttt', (message, topic) => {
            console.log(`\nNova mensagem no tópico ${topic}:`, message);
        });

        // 4. Publicar uma mensagem
        let cont = 0
        await client.publish('testtttrttt', 'Olá do Node.js!' + cont);
        setInterval(async () => {
            cont++
            await client.publish('testtttrttt', 'Olá do Node.js!' + cont);
        }, 200);



        // console.log('\nAguardando mensagens... (Ctrl+C para sair)');
        //  await new Promise(resolve => setTimeout(resolve, 60000));

    } catch (err) {
        console.error('Erro:', err.message);
    } finally {
        // 7. Desconectar
        //   client.disconnect();
        console.log('Conexão encerrada.');
    }
}

// Executar o exemplo
main().catch(console.error);