package api

import (
	"encoding/json"
	"log"
	"osone/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zishang520/socket.io/v2/socket"
)

type Body struct {
	Service string            `json:"service"`
	Data    map[string]string `json:"data"`
}

func (b *Body) Str() string {
	by, _ := json.Marshal(b.Data)
	return string(by)
}

func JWTMiddlewarevService() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apikey := c.Get("apikey")
		if apikey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token não fornecido"})
		}
		str, err := utils.Decrypt(apikey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token inválido"})
		}

		c.Locals("service", str)
		return c.Next()
	}
}
func (s *Server) configRouteVirtial() {

	s.app.Post("/api/services", JWTMiddlewarevService(), func(c *fiber.Ctx) error {
		jsonStr := c.Body()
		codeVal := c.Locals("service")
		if codeVal == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not found service"})
		}

		service := map[string]string{}
		err := json.Unmarshal([]byte(codeVal.(string)), &service)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		input := Body{}
		err = json.Unmarshal([]byte(jsonStr), &input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := s.mq.Request(service["virtual"], input.Service, input.Str(), 10*time.Second)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		jsonres := map[string]interface{}{}
		err = json.Unmarshal([]byte(res), &jsonres)
		if err != nil {
			return c.JSON(res)
		}
		return c.JSON(jsonres)
	})
}

func (s *Server) configWsVirtial() {
	s.io.Of("/virtual", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)
		//next := clients[1].(func(*socket.ExtendedError))

		handshake := client.Handshake() // Obtemos o objeto Handshake diretamente
		if handshake.Auth == nil {
			client.Disconnect(true)
			return
		}
		auth := handshake.Auth.(map[string]interface{})

		verify := map[string]func(token string) (map[string]string, error){}
		verify["jwt"] = func(token string) (map[string]string, error) {

			res := map[string]string{"test": "dddd"}
			return res, nil
		}
		verify["apikey"] = func(token string) (map[string]string, error) {

			str, err := utils.Decrypt(token)
			if err != nil {
				return nil, err
			}
			res := map[string]string{}
			err = json.Unmarshal([]byte(str), &res)
			if err != nil {
				return nil, err
			}
			return res, nil

		}

		if verify[auth["type"].(string)] == nil {

			client.Disconnect(true)
			return
		}

		data, err := verify[auth["type"].(string)](auth[auth["type"].(string)].(string))
		if err != nil {
			client.Disconnect(true)
			return
		}
		client.SetData(data)

	}).On("connection", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)

		if client.Data() == nil {

			client.Disconnect(true)
			return
		}
		auth := client.Data().(map[string]string)

		log.Println("Novo usuário conectado:", client.Id()) // Corrigido para 'Id'
		virtual := auth["virtual"]
		// Evento de recebimento de mensagens
		client.On("mq:subscribe", func(args ...interface{}) {
			topic_, ok := args[0].(string)
			if !ok {

				return
			}

			s.mq.Subscribe(virtual, topic_, func(message, topic string) {
				client.Emit("mq:"+topic_, message, topic)
			})
		})

		client.On("mq:publish", func(args ...interface{}) {

			topic, ok := args[0].(string)
			if !ok {
				return
			}
			message, ok := args[1].(string)
			if !ok {
				return
			}
			s.mq.Publish(virtual, topic, message)
		})

		///
		client.On("mqtt:subscribe", func(args ...interface{}) {
			topic_, ok := args[0].(string)
			if !ok {

				return
			}

			s.mqtt.Subscribe(virtual+"/"+topic_, func(message, topic string) {
				t := strings.ReplaceAll(topic, virtual+"/", "")
				client.Emit("mqtt:on", message, t)
			}, 1)
		})

		client.On("mqtt:publish", func(args ...interface{}) {

			topic, ok := args[0].(string)
			if !ok {
				return
			}
			message, ok := args[1].(string)
			if !ok {
				return
			}
			s.mqtt.Publish(virtual+"/"+topic, message)
		})
		// Evento de desconexão
		client.On("disconnect", func(args ...interface{}) {
			log.Println("Usuário desconectado:", client.Id()) // Corrigido para 'Id'
		})
	})

}
