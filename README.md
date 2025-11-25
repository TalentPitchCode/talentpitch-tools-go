# TalentPitch Tools Go

Herramientas compartidas para proyectos Go de TalentPitch, incluyendo middlewares para Gin.

## Instalación

```bash
go get github.com/TalentPitchCode/talentpitch-tools-go@latest
```

## Uso

### Configurar Middlewares Básicos (sin JWT)

```go
import (
    "github.com/gin-gonic/gin"
    talentpitchtools "github.com/TalentPitchCode/talentpitch-tools-go"
)

func main() {
    router := gin.New()
    
    // Setup middlewares without JWT
    router, err := talentpitchtools.SetupTalentpitchMiddlewaresWithoutJWT(router)
    if err != nil {
        log.Fatal(err)
    }
    
    // Ahora puedes usar c.GetString("client_ip") en tus handlers
    router.GET("/ip", func(c *gin.Context) {
        ip := c.GetString("client_ip")
        c.JSON(200, gin.H{"ip": ip})
    })
}
```

### Configurar Middlewares con JWT

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
    talentpitchtools "github.com/TalentPitchCode/talentpitch-tools-go"
)

// Define tus custom claims
type CustomClaims struct {
    jwt.StandardClaims
    UserID   uint   `json:"user_id"`
    Email    string `json:"email"`
    // ... otros campos
}

func (c CustomClaims) Valid() error {
    // Implementa tu validación
    return c.StandardClaims.Valid()
}

func main() {
    router := gin.New()
    jwtSecret := "tu-secret-key"
    
    // Crea un factory para tus claims
    claimsFactory := func() jwt.Claims {
        return &CustomClaims{}
    }
    
    // Setup middlewares with JWT
    router, err := talentpitchtools.SetupTalentpitchMiddlewares(router, jwtSecret, claimsFactory)
    if err != nil {
        log.Fatal(err)
    }
    
    // Middleware JWT requerido
    router.Use(talentpitchtools.JWTMiddleware(jwtSecret, claimsFactory))
}
```

## Características

### Client IP Middleware

El middleware `clientIPMiddleware` calcula automáticamente la IP real del cliente desde:
- `X-Forwarded-For` header (ngrok, Cloudflare, etc.)
- `X-Real-IP` header
- `c.ClientIP()` como fallback

La IP se guarda en el contexto y puedes accederla con:
```go
ip := c.GetString("client_ip")
```

### Location Middleware

Configura automáticamente el esquema y host desde los headers del proxy.

### JWT Middleware

Middlewares opcionales y requeridos para validación JWT con soporte para custom claims.

## Requisitos

- Go 1.23+
- Gin v1.10.0+

## Licencia

Propietario - TalentPitch

