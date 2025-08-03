package api

import (
	"fmt"
	"osone/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zishang520/socket.io/v2/socket"
)

type AuthLogin struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// Função para verificar token de autenticação
func verifyToken(token string) bool {
	// Aqui você pode adicionar a lógica para validar o token, por exemplo, verificando em um banco de dados ou decodificando um JWT
	return token == "valid_token" // Exemplo simples, substitua pela lógica real
}

func JWTMiddleware(roles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token não fornecido"})
		}

		// Remover o prefixo "Bearer " caso exista
		parts := strings.Split(tokenString, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			tokenString = parts[1]
		}

		claims, ok, err := utils.VerifyJWT(tokenString)

		if err != nil || !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token inválido"})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

func JWTMiddlewarev1(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token não fornecido"})
		}

		// Remover o prefixo "Bearer " caso exista
		parts := strings.Split(tokenString, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			tokenString = parts[1]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Método de assinatura inválido")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token inválido"})
		}

		// Definir a claim no contexto para acesso posterior
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Falha ao processar claims"})
		}
		c.Locals("user", claims)
		return c.Next()
	}
}

// Middleware para autenticação via token
func authenticateSocket(client *socket.Socket, next func(*socket.ExtendedError)) {

	verify := map[string]func(token string) (map[string]string, error){}
	verify["jwt"] = func(token string) (map[string]string, error) {
		res := map[string]string{"test": "dddd"}
		return res, nil
	}
	verify["apikey"] = func(token string) (map[string]string, error) {
		res := map[string]string{"test": "ddddd"}
		return res, nil
	}
	handshake := client.Handshake() // Obtemos o objeto Handshake diretamente

	if handshake.Auth == nil {
		next(socket.NewExtendedError("not authorized", ""))
		client.Disconnect(true)
		return
	}
	auth := handshake.Auth.(map[string]string)

	if verify[auth["type"]] == nil {
		next(socket.NewExtendedError("not authorized", ""))
		client.Disconnect(true)
		return
	}

	data, err := verify[auth["type"]](auth[auth["type"]])
	if err != nil {
		next(socket.NewExtendedError("not authorized", ""))
		client.Disconnect(true)
		return
	}

	client.SetData(data)
	fmt.Println(data, "data", client.Data())
	next(nil)
	/*

		auth, exists := handshake.Auth.(map[string]interface{})
		if !exists {
			log.Println("Campo 'auth' ausente no handshake")
			client.Disconnect(true)
			return
		}

		token, tokenExists := auth["token"].(string)
		if !tokenExists || !verifyToken(token) {
			log.Println("Token inválido ou ausente")
			client.Disconnect(true)
			return
		}

		next(nil)
	*/
}

func (s *Server) configRouterAuth() {
	/*

		s.app.Post("/api/users", JWTMiddleware([]string{"admin"}), func(c *fiber.Ctx) error {
			jsonStr := c.Body()
			var input services.User
			err := json.Unmarshal([]byte(jsonStr), &input)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			userId, err := s.service.CreateUser(&input)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(fiber.Map{"id": userId})
		})

		s.app.Get("/api/auth/verify/:identifier", func(c *fiber.Ctx) error {

			identifier := c.Params("identifier")
			user, err := s.service.GetUserByCode(identifier)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(fiber.Map{"code": user.Code})
		})
		s.app.Post("/api/auth/verify", func(c *fiber.Ctx) error {
			jsonStr := c.Body()
			fmt.Println(jsonStr)
			var input services.UserAuth

			err := json.Unmarshal([]byte(jsonStr), &input)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}

			fmt.Println(input)
			user, err := s.service.GetUserVerify(input.Code, input.CodeAuth)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(user)
		})

		s.app.Get("/api/auth/jwt", JWTMiddleware([]string{"admin"}), func(c *fiber.Ctx) error {
			user := c.Locals("user")
			return c.JSON(user)
		})
		////////////////////////////////////////////
		s.app.Post("/api/auth/login", func(c *fiber.Ctx) error {
			jsonStr := c.Body()
			var input AuthLogin
			err := json.Unmarshal([]byte(jsonStr), &input)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			user, err := s.service.GetUserLogin(input.Identifier, input.Password)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(user)
		})

	*/

}
