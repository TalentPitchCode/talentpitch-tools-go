package talentpitchtools

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/TalentPitchCode/talentpitch-tools-go/helpers"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

// SetupLocationWithTrustedProxies configures Gin router with location middleware
// and trusted proxies settings. This function should be called before setting up routes.
func SetupLocationWithTrustedProxies(r *gin.Engine, jwtSecret string, trustedProxies []string) (*gin.Engine, error) {
	// Trust all proxies (Required for Cloudflare -> AWS ALB -> EKS)
	// Security is handled by AWS Security Groups and VPC isolation
	// Ingress: Tu ALB (k8s-developm-nginx...) tiene los Security Groups sg-087e406bb9c504ccf y sg-00191405ecc229d51.
	// Nodos: Tus nodos usan el Security Group sg-0ea1c17719c2b71f6.
	// Reglas:
	// ✅ Existe una regla que permite tráfico desde el ALB hacia tus nodos en el rango de puertos 3030-5678 (que incluye tu puerto 5001).
	// ✅ NO existen reglas de entrada abiertas (0.0.0.0/0) en tus nodos.
	// Conclusión: Nadie puede conectarse directamente a tus Pods desde internet saltándose el Load Balancer. Por lo tanto, confiar en todas las IPs (0.0.0.0/0) a nivel de aplicación es seguro porque la red ya filtra quién puede hablarte (solo el ALB).
	r.SetTrustedProxies(trustedProxies)

	// Use location middleware (handles scheme/host from headers)
	// Note: c.ClientIP() should work automatically after SetTrustedProxies
	r.Use(location.Default())

	// Use ClientIP middleware to calculate and store client IP in context
	r.Use(clientIPMiddleware())

	// Use JWT middleware if jwtSecret is provided
	if jwtSecret != "" {
		r.Use(optionalJWTMiddleware(jwtSecret))
	}

	return r, nil
}

/*****************************************************************
* Function Name: OptionalJWTMiddleware
* Description: Optional middleware for JWT validation
* If token is present and valid, sets user in context
* If token is missing or invalid, continues without setting user
*****************************************************************/
func optionalJWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		tokenSplit := strings.Split(tokenHeader, " ")
		if len(tokenSplit) != 2 {
			// Invalid token format, continue without authentication
			c.Next()
			return
		}

		tokenString := tokenSplit[1]
		//token validation
		token, err := jwt.ParseWithClaims(tokenString, &helpers.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// If token is valid, set user in context
		claims := token.Claims.(*helpers.CustomClaims)
		c.Set("user", claims)

		c.Next()
	}
}

/*****************************************************************
* Function Name: JWTMiddleware
* Description: Middleware for validate JWT (required authentication)
*****************************************************************/
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenSplit := strings.Split(tokenHeader, " ")
		if len(tokenSplit) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := tokenSplit[1]
		//token validation
		token, err := jwt.ParseWithClaims(tokenString, &helpers.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		// if token is valid, set user in context
		claims := token.Claims.(*helpers.CustomClaims)

		c.Set("user", claims)

		c.Next()
	}
}

func SwaggerBasicAuth(email, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		email: password,
	})
}

/*****************************************************************
* Function Name: clientIPMiddleware
* Description: Middleware that calculates client IP and stores it in context
* Usage: router.Use(talentpitchtools.clientIPMiddleware())
* Then use: c.GetString("client_ip") to get the IP
*****************************************************************/
func clientIPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		c.Set("client_ip", ip)
		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (used by ngrok, Cloudflare, etc.)
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// The first IP is the original client IP
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			// Validate it's a valid IP address
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header (alternative header used by some proxies)
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		ip := strings.TrimSpace(realIP)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Fallback to Gin's ClientIP() method
	return c.ClientIP()
}

// SetupTalentpitchMiddlewares is a convenience function that sets up all middlewares
func SetupTalentpitchMiddlewares(r *gin.Engine, jwtSecret string, trustedProxies []string) (*gin.Engine, error) {
	return SetupLocationWithTrustedProxies(r, jwtSecret, trustedProxies)
}

