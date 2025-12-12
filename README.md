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
    
    // Ahora puedes usar c.GetString("client_ip") y c.MustGet("user")
    router.GET("/me", func(c *gin.Context) {
        // Note: This example assumes a JWT middleware has run and populated "user" context key,
        // even though this section is titled "sin JWT".
        // For a full JWT setup, refer to the "Configurar Middlewares con JWT" section.
        user := c.MustGet("user").(*helpers.CustomClaims)
        ip := c.GetString("client_ip")
        c.JSON(200, gin.H{"user": user, "ip": ip})
    })
    
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

### GROQ Message Filtering

El paquete incluye funcionalidad para filtrar mensajes usando GROQ AI. El paquete es público y no inyecta variables directamente, pero puede leer variables de entorno de los proyectos que lo usan.

#### Configuración

El paquete lee las siguientes variables de entorno (o puede configurarse programáticamente):

- `GROQ_API_KEY`: Tu API key de Groq (requerido)
- `GROQ_MODEL`: Modelo de Groq a usar (opcional, por defecto: "llama-3.1-8b-instant")

#### Uso Básico

```go
import (
    "context"
    "time"
    
    "github.com/TalentPitchCode/talentpitch-tools-go/groq"
)

func main() {
    // Inicializar cliente Groq (lee variables de entorno automáticamente)
    groqClient := groq.NewClient(groq.Config{})
    
    if groqClient == nil {
        log.Fatal("Failed to initialize Groq client. Make sure GROQ_API_KEY is set.")
    }
    
    // Verificar contenido de mensaje
    ctx := context.Background()
    messageText := "Hello, this is a test message"
    
    isMalicious, errorCode, reason, err := groqClient.CheckMessageContent(ctx, messageText)
    if err != nil {
        log.Printf("Error checking message: %v", err)
        // Fail open - permitir mensaje si hay error
    } else if isMalicious {
        log.Printf("Message rejected: %s - %s", errorCode, reason)
        // Rechazar mensaje
    } else {
        // Mensaje seguro, proceder
    }
}
```

#### Uso con Filtrado de Mensajes

```go
import (
    "context"
    "time"
    
    "github.com/TalentPitchCode/talentpitch-tools-go/groq"
)

func main() {
    groqClient := groq.NewClient(groq.Config{})
    
    ctx := context.Background()
    messageText := "Check this message"
    
    // Usar FilterMessageWithAI (alias de CheckMessageContent)
    isMalicious, errorCode, reason, err := groqClient.FilterMessageWithAI(ctx, messageText)
    if err != nil {
        // Manejar error
    } else if isMalicious {
        // Mensaje malicioso detectado
    }
}
```

#### Guardar Mensajes Maliciosos

Para guardar mensajes rechazados en tu propia base de datos, implementa la interfaz `MaliciousMessageSaver`:

```go
import (
    "time"
    
    "github.com/TalentPitchCode/talentpitch-tools-go/groq"
    "gorm.io/gorm"
)

// Implementar la interfaz MaliciousMessageSaver
type MyMaliciousMessageSaver struct {
    DB *gorm.DB
}

func (s *MyMaliciousMessageSaver) SaveMaliciousMessage(
    fromUserID int, 
    toUserID int, 
    messageText string, 
    errorCode string, 
    reason string, 
    currentTime string,
) error {
    maliciousMessage := MaliciousMessageTable{
        FromUserID: fromUserID,
        ToUserID:   toUserID,
        Message:    messageText,
        ErrorCode:  errorCode,
        Reason:     reason,
        CreatedAt:  currentTime,
        UpdatedAt:  currentTime,
    }
    
    return s.DB.Table("blocked_malicious_messages").Create(&maliciousMessage).Error
}

// Uso en tu código
func filterAndSaveMessage(groqClient *groq.Client, saver groq.MaliciousMessageSaver, fromUserID, toUserID int, messageText string) error {
    ctx := context.Background()
    
    // Filtrar mensaje
    isMalicious, errorCode, reason, err := groqClient.FilterMessageWithAI(ctx, messageText)
    if err != nil {
        return err
    }
    
    if isMalicious {
        // Guardar mensaje malicioso
        loc, _ := time.LoadLocation("America/Bogota")
        currentTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
        
        if err := groq.SaveMaliciousMessage(saver, fromUserID, toUserID, messageText, errorCode, reason, currentTime); err != nil {
            log.Printf("Error saving malicious message: %v", err)
        }
        
        return fmt.Errorf("message rejected: %s", errorCode)
    }
    
    return nil
}
```

#### Configuración Programática

También puedes configurar el cliente programáticamente en lugar de usar variables de entorno:

```go
groqClient := groq.NewClient(groq.Config{
    APIKey: "your-api-key-here",
    Model:  "llama-3.1-8b-instant",
    BaseURL: "https://api.groq.com/openai/v1", // Opcional, tiene valor por defecto
})
```

#### Personalizar el Prompt

Puedes inyectar tu propio prompt template en el setup del cliente:

```go
import (
    "fmt"
    "github.com/TalentPitchCode/talentpitch-tools-go/groq"
)

// Definir tu prompt personalizado
customPrompt := func(messageText string) string {
    return fmt.Sprintf(`Tu prompt personalizado aquí.

Mensaje: "%s"

Responde con un JSON en este formato:
{
  "is_malicious": true o false,
  "error_code": "ERROR_CODE" o null,
  "reason": "razón breve"
}`, messageText)
}

// Crear cliente con prompt personalizado
groqClient := groq.NewClient(groq.Config{
    PromptTemplate: customPrompt,
    // ... otras configuraciones
})
```

El prompt se usará automáticamente en `CheckMessageContent` y `FilterMessageWithAI`. Si no proporcionas un `PromptTemplate`, se usará el prompt por defecto.

#### Códigos de Error

El filtro puede retornar los siguientes códigos de error:

- `CONTENT_SPAM`: Mensajes spam
- `CONTENT_INAPPROPRIATE`: Contenido inapropiado
- `CONTENT_HARASSMENT`: Acoso o bullying
- `CONTENT_SCAM`: Estafas o phishing
- `CONTENT_VIOLENCE`: Contenido violento o amenazante
- `CONTENT_OTHER`: Otro contenido malicioso

## Requisitos

- Go 1.23+
- Gin v1.10.0+ (para middlewares)
- Variable de entorno `GROQ_API_KEY` (para filtrado de mensajes)

## Licencia

Propietario - TalentPitch

