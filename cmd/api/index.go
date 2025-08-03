package api

import (
	"fmt"
	"log"
	"os/signal"
	"osone/utils"
	"syscall"
	"time"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

type Server struct {
	app  *fiber.App
	io   *socket.Server
	mq   *utils.MQbroker
	mqtt *utils.MQTTbroker
}

func (s *Server) configureWebSocket() {
	c := socket.DefaultServerOptions()
	c.SetConnectionStateRecovery(&socket.ConnectionStateRecovery{})
	c.SetAllowEIO3(true)
	c.SetPingInterval(300 * time.Millisecond)
	c.SetPingTimeout(200 * time.Millisecond)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(1000 * time.Millisecond)
	c.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	// Adiciona middleware de autenticação
	s.io.Use(authenticateSocket)

	// Manipulador de eventos de conexão
	s.io.On("connection", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)
		log.Println("Novo usuário conectado:", client.Id) // Corrigido para 'Id'

		// Evento de recebimento de mensagens
		client.On("message", func(args ...interface{}) {
			message, ok := args[0].(string)
			if !ok {
				log.Println("Mensagem inválida recebida")
				return
			}
			log.Printf("Mensagem recebida: %s", message)
			client.Broadcast().Emit("message", message)
		})

		// Evento de desconexão
		client.On("disconnect", func(args ...interface{}) {
			log.Println("Usuário desconectado:", client.Id) // Corrigido para 'Id'
		})
	})

	s.configWsVirtial()
}

func (s *Server) configureFiberApp() {

	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Permitir todas as origens, pode ser alterado para domínios específicos
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	// Servindo arquivos estáticos
	s.app.Static("/", "public")

	// Rotas do Socket.IO
	s.app.Use("/socket.io/", adaptor.HTTPHandler(s.io.ServeHandler(nil)))
	s.app.Put("/socket.io", adaptor.HTTPHandler(s.io.ServeHandler(nil)))
	s.app.Get("/socket.io", adaptor.HTTPHandler(s.io.ServeHandler(nil)))
	s.app.Post("/socket.io", adaptor.HTTPHandler(s.io.ServeHandler(nil)))

	// Wildcard - greedy - optional
	s.app.Get("/api/apps/:appId/processes/:pId/*", func(c *fiber.Ctx) error {
		fmt.Println(c.Params("appId"), c.Params("pId"))
		return c.SendString(c.Params("*"))
	})
	s.configRouteVirtial()
	/*
		s.configRouterApp()
		s.configRouterTrailer()
		s.configRouterAuth()
		s.configRouterTempate()
		s.configRouterUser()
	*/
	/*
		s.app.Use(func(c *fiber.Ctx) error {
			return c.SendFile("./public/index.html")
		})

	*/
	s.app.Use(func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})

}

func (s *Server) Run(url string) {

	s.configureWebSocket()
	//s.deviceWebSocket()
	s.configureFiberApp()

	// Iniciando o servidor Fiber em uma goroutine
	go func() {
		if err := s.app.Listen(url); err != nil {
			log.Fatalf("Erro ao iniciar o servidor Fiber: %v", err)
		}
	}()

	log.Println("Servidor iniciado na " + url)

	// Captura sinais do sistema para encerramento seguro
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-exit
	log.Println("Sinal de encerramento recebido, fechando o servidor...")

	// Encerra o servidor Socket.IO
	s.io.Close(nil)

	// Encerra o servidor Fiber
	s.app.Shutdown()

	log.Println("Servidor encerrado com sucesso")
}

func NewServer(mq *utils.MQbroker, mqtt *utils.MQTTbroker) *Server {

	return &Server{
		mq:   mq,
		mqtt: mqtt,
		io:   socket.NewServer(nil, nil),
		app:  fiber.New(),
	}

}
