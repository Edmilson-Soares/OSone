import { io } from "socket.io-client"

// Altere para o endereço real do seu servidor
const socket = io("http://localhost:3333", {
    transports: ["websocket"], // Evita fallback para long polling
    auth: {
        mqtt: {
            username: "device",
            password: "HHGJDHDJDKDJHDJBGJDBJUHBDJBDJBDBJDBJDBJBDJDB"
        },
        mq: {
            jwt: "mq://root:fffffffffffffffffff@127.0.0.1:4051"
        }
    }
})

socket.on("connect", () => {
    console.log("✅ Conectado ao servidor com id:", socket.id)

})
socket.on("mqtt:connect", (data) => {
    console.log("✅ Conectado ao MQTT")
    socket.emit("mqtt:subscribe", "teste/#", () => {
        setInterval(() => {
            socket.emit("mqtt:publish", "teste/topico", "Olá do Bun MQTT client!");
        }, 500)
    })

})
socket.on("mqtt:message", (topic, message) => {
    console.log(topic, message)
})

socket.on("disconnect", () => {
    console.log("❌ Desconectado do servidor")
})



socket.on("mq:connect", (id) => {
    console.log("✅ Conectado ao MQ", id)
    socket.emit("mq:subscribe", "teste.topico");
    setInterval(() => {
        socket.emit("mq:publish", "teste.topico", "Olá do Bun MQ client!");
    }, 500)


})

socket.on("mq:on:teste.topico", (data) => {
    console.log(data.topic, data.payload)
})
socket.on("mq:err", (err) => {
    console.log(err)
})