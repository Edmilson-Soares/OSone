import { io } from "socket.io-client"

const delay=(ms)=>new Promise((resolve, reject) => {
    setTimeout(() => {  
        resolve()
        }, ms)
        
})

// Altere para o endereço real do seu servidor
const socket = io("http://localhost:1337/virtual", {
   // transports: ["websocket"], // Evita fallback para long polling
    auth: {
        type:"apikey",
        apikey: "8rQJ0Afl1sY3NfC57b67X4sgrciuUoVWFtKO4IZQSH_4HsfkVb8yYJAXfCF_nyJTdIpG0C1_IcCfG_Jn_5fJzj2W5osg3A90nPnq3cMi2Oysj8SNU_yuNSDfVw=="
    }
})

socket.on("connect", () => {
    console.log("✅ Conectado ao servidor com id:", socket.id)

})
socket.on("mqtt:connect", (data) => {
    console.log("✅ Conectado ao MQTT")


})


await delay(1000)
    socket.emit("mq:subscribe","test.*.test")
    socket.emit("mqtt:subscribe","test")
    socket.on("mq:test.*.test", (data) => {
        console.log(data)
    })

    socket.on("mqtt:on", (message,topic) => {
        console.log(message,topic,"mqtt")
    })
await delay(1000)
 //socket.emit("mq:publish","test.test.test","tests")
 socket.emit("mqtt:publish","test","ws---")
socket.on("disconnect", () => {
    console.log("❌ Desconectado do servidor")
})
socket.on("error", (error) => {
    console.log("❌ Erro:", error)
}
)