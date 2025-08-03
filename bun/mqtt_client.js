import mqtt from 'mqtt';

// Configurações de autenticação
const options = {
  username: 'admin',
  password: 'admin'
};
//37.27.39.202
//localhost
// Publicador
{
  const client = mqtt.connect('mqtt://37.27.39.202:1883', options);

  client.on('connect', () => {
    console.log('📤 Conectado ao broker MQTT como publicador.');

    setInterval(() => {
      const mensagem = `Olá MQTT! ${new Date().toLocaleTimeString()}`;
      client.publish('topico/exemplo', mensagem);
      console.log(`📤 Mensagem publicada: ${mensagem}`);
    }, 3000);
  });
}

// Assinante
{
  const client = mqtt.connect('mqtt://37.27.39.202:1883', options);

  client.on('connect', () => {
    console.log('📥 Conectado ao broker MQTT como assinante.');
    client.subscribe('topico/#', (err) => {
      if (!err) {
        console.log('✅ Subscreveu ao tópico: topico/exemplo');
      }
    });
  });

  client.on('message', (topic, message) => {
    console.log(`📩 Mensagem recebida em '${topic}': ${message.toString()}`);
  });
}
