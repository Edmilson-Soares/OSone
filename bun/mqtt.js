import mqtt from 'mqtt';

// Configurações de autenticação
const options = {
  username: 'test',
  password: '5253b332-70a5-4749-ad5d-d79f0e689d79'
};
//37.27.39.202
//localhost
// Publicador
  const client = mqtt.connect('mqtt://localhost:1883', options);

  client.on('connect', () => {
    console.log('📤 Conectado ao broker MQTT como publicador.');
    client.subscribe('/test/test/#', (err) => {
      if (!err) {
        console.log('✅ Subscreveu ao tópico: topico/exemplo');
      }
    });
    setInterval(() => {
      const mensagem = `Olá MQTT! ${new Date().toLocaleTimeString()}`;
      client.publish('/test/test/exemplo', mensagem);
     // console.log(`📤 Mensagem publicada: ${mensagem}`);
    }, 3000);
  });



  client.on('message', (topic, message) => {
    console.log(`📩 Mensagem recebida em '${topic}': ${message.toString()}`);
  });

  /*
const options = {
  username: 'apikey',
  password: 'R7cy_ZOLwqgchqMEd0cA8Vo5Tc4oKoHzVnte7o32lH9BVupHQ7fkAERlaEuQyjGgc32ANDB2KvJnC9axB33JAX35qckQxxyWRm7wJlzSfvVTn05a-Dtf8124SOZ-NoIQ07MkVrVf0Xfvh-wW_yYHa13tmusUSpUF5D8JBi7-3_VWQWz3BZHwW7GfiBU_xWoxVU_cbgfbq8WTpkspH3_Jc53ofK1Lc4ogsh2CIcCqnMfrYvdiADBcAig0AnalMuN7jGhM-oOsKz-HJHy0rM1BoaYVgARSPvoqQf6Plm-k7u-8U7TsPzqhm6UZ7z684eXCvvs_3sJiV_Hdmkm60qAE1Rt1uGktE2EEaePPFQEz8UTi7koWiKM='
};
//37.27.39.202
//localhost
// Publicador
  const client = mqtt.connect('mqtt://localhost:1883', options);

  client.on('connect', () => {
    console.log('📤 Conectado ao broker MQTT como publicador.');
    client.subscribe('/test/#', (err) => {
      if (!err) {
        console.log('✅ Subscreveu ao tópico: topico/exemplo');
      }
    });
    setInterval(() => {
      const mensagem = `Olá MQTT! ${new Date().toLocaleTimeString()}`;
      client.publish('/test/exemplo', mensagem);
     // console.log(`📤 Mensagem publicada: ${mensagem}`);
    }, 3000);
  });



  client.on('message', (topic, message) => {
    console.log(`📩 Mensagem recebida em '${topic}': ${message.toString()}`);
  });
\"username\":\"test\",\"passworrd\":\"5253b332-70a5-4749-ad5d-d79f0e689d79\"
  */