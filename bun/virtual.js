import { MQ }  from './mq-client'; // Assumindo que o código está no arquivo mq-client.js

async function main() {
    // 1. Criar uma instância do cliente MQ
    const client = new MQ();

    try {
        // 2. Conectar ao servidor MQ
          await client.connect('mq://root:chave-secreta-32-bytes-123456789@localhost:4222');
        console.log('Conectado com sucesso! ID:', client.ID);

       /*
        const response = await client.request('osone.addVirtual', {
             name:"test",
             desc:"ddddd",
             enterpriseId:"deb52666-535f-42ca-9d9f-a7566946692e"

        }, 3000);
        
        console.log('\nResposta:', response);
       //"dc4fc57c-5a6f-4ff1-8e3b-7184262a3770"
     const response = await client.request('osone.editVirtual', {
             id:"dc4fc57c-5a6f-4ff1-8e3b-7184262a3770",
             desc:"fffffffffffffffff",
            icon:"ddddd",

        }, 3000);
              console.log('\nResposta:', response);

        const response = await client.request('osone.delVirtual', {
             id:"dc4fc57c-5a6f-4ff1-8e3b-7184262a3770",
        }, 3000);
        console.log('\nResposta:', response);



                const response = await client.request('osone.getVirtual',"ebd39893-60cb-4ca9-90b5-b254de1c480d", 3000);
        
        console.log('\nResposta:', response);
        */

        /*
   
        const response = await client.request('osone.addDevice', {
             name:"test",
             code:"test",
             desc:"ddddd",
             network:{
                ip:"192.168.1.1",
                mac:"00:11:22:33:44:55",
                port: "8080"

             },
             config:{
                "key1":"value1",
                "key2":"value2",

             },
             location:{
                latitude: "-23.5505",
                longitude: "-46.6333",
             },
             virtualId:"ebd39893-60cb-4ca9-90b5-b254de1c480d"

        }, 3000);
          console.log('\nResposta:', response);


             
        const response = await client.request('osone.editDevice', {
             id: "fe9348c9-8a53-403d-a654-c80c383a4765",
             name:"test",
             code:"test",
             desc:"ddddd",

        }, 3000);
          console.log('\nResposta:', response);


        const response = await client.request('osone.networkDevice', {
            id: "fe9348c9-8a53-403d-a654-c80c383a4765",
            ip:"192.168.1.1",
            mac:"00:11:22:33:44:55",
            port: "8080"
        }, 3000);
        console.log('\nResposta:', response);


        
        const response = await client.request('osone.locationDevice', {
            id: "fe9348c9-8a53-403d-a654-c80c383a4765",
            latitude: "-23.5505",
            longitude: "-46.6333",
        }, 3000);
        console.log('\nResposta:', response);


        const response = await client.request('osone.locationDevice', {
            id: "fe9348c9-8a53-403d-a654-c80c383a4765",
            "key1":"value1",
            "key2":"value2",
        }, 3000);
        console.log('\nResposta:', response);


        const response = await client.request('osone.authDevice', {
            id: "fe9348c9-8a53-403d-a654-c80c383a4765",
        }, 3000);
        console.log('\nResposta:', response);


        const response = await client.request('osone.delDevice', {
            id: "fe9348c9-8a53-403d-a654-c80c383a4765",
        }, 3000);
        console.log('\nResposta:', response);

          */


     const response = await client.request('osone.getVirtual',"ebd39893-60cb-4ca9-90b5-b254de1c480d", 3000);
        
        console.log('\nResposta:', response);
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