# GMF Message Processor

GMF Message Processor es una aplicación desarrollada en Go que procesa mensajes recibidos a través de Amazon SQS,
realiza operaciones con plantillas almacenadas en una base de datos Postgres y envía correos electrónicos utilizando
SMTP o Amazon SES. La aplicación está diseñada para ser escalable y cumplir con las mejores prácticas de arquitectura y
desarrollo en Go.

## Tabla de Contenidos

- [Características](#características)
- [Tecnologías Usadas](#tecnologías-usadas)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Configuración Inicial](#configuración-inicial)
- [Uso](#uso)
- [Pruebas](#pruebas)
- [Despliegue](#despliegue)
- [Contribuir](#contribuir)
- [Licencia](#licencia)

## Características

- Procesa mensajes provenientes de una cola de Amazon SQS.
- Administra plantillas de correo electrónico almacenadas en una base de datos Postgres.
- Soporte para el envío de correos utilizando:
    - **Amazon SES** (para entornos de producción).
    - **SMTP** (para entornos locales).
- Manejo de secretos a través de AWS Secrets Manager o variables de entorno.
- Uso de principios SOLID y mejores prácticas en la arquitectura de Go.
- Fácilmente configurable mediante variables de entorno.

## Tecnologías Usadas

- **Go**: Lenguaje de programación principal.
- **Amazon Web Services (AWS)**:
    - **Amazon SES**: Para el envío de correos.
    - **Amazon SQS**: Para la cola de mensajes.
    - **AWS Secrets Manager**: Para la gestión de secretos como credenciales.
- **Postgres**: Base de datos para almacenar plantillas de correos.
- **Logrus**: Librería de logging para registrar eventos.
- **Gorm**: ORM para interactuar con la base de datos.
- **Testify**: Para escribir pruebas unitarias y mocks.
- **Docker**: Para la creación de entornos reproducibles.

## Estructura del Proyecto

El proyecto sigue una estructura modular clara y organizada:

    ```plaintext

├── bootstrap
├── cmd
│ └── lambda
│ └── main.go
├── config
│ ├── config.go
│ ├── config_manager_test.go
│ └── init.go
├── connection
│ ├── db_manager.go
│ ├── db_manager_test.go
│ └── get_secret.go
├── coverage.html
├── coverage.out
├── function.zip
├── go.mod
├── go.sum
├── internal
│ ├── aws
│ │ ├── sqs_client.go
│ │ └── sqs_client_test.go
│ ├── email
│ │ ├── ses.go
│ │ ├── smtp.go
│ │ └── smtp_test.go
│ ├── handler
│ │ ├── handler_test.go
│ │ └── sqs_handler.go
│ ├── logs
│ │ ├── logger.go
│ │ └── logger_test.go
│ ├── models
│ │ ├── model_test.go
│ │ ├── plantilla.go
│ │ └── sqs_message.go
│ ├── repository
│ │ ├── repository.go
│ │ └── repository_test.go
│ ├── service
│ │ ├── service.go
│ │ └── service_test.go
│ └── utils
│ ├── sqs_helper.go
│ ├── sqs_interface.go
│ ├── text_replacer.go
│ ├── utils_test.go
│ └── validate_message.go
├── README.md
├── seeds
│ ├── data
│ │ └── plantillas.json
│ └── load_data_plantilla.go
├── sonar-project.properties
└── test_data
├── message.json
└── no_messages.json

    ```

## Variables de Entorno

La aplicación se configura mediante variables de entorno. A continuación se muestra una lista de las variables
necesarias para ejecutar la aplicación:

- **DB_HOST**: Host de la base de datos.
- **DB_PORT**: Puerto de la base de datos.
- **DB_NAME**: Nombre de la base de datos.
- **DB_USER**: Usuario de la base de datos.
- **DB_PASSWORD**: Contraseña de la base de datos.
- **SMTP_HOST**: Host del servidor SMTP.
- **SMTP_PORT**: Puerto del servidor SMTP.
- **SMTP_USER**: Usuario del servidor SMTP.
- **SMTP_PASSWORD**: Contraseña del servidor SMTP.
- **SQS_QUEUE_URL**: URL de la cola de mensajes de Amazon SQS.
- **SECRETS_DB**: Nombre del secreto en AWS Secrets Manager que contiene las credenciales de la base de datos.
- **SECRETS_SMTP**: Nombre del secreto en AWS Secrets Manager que contiene las credenciales del servidor SMTP.

## Instalacion de dependencias

Para instalar las dependencias del proyecto, ejecute el siguiente comando:

```bash
go mod tidy
```

# Uso

Para ejecutar la aplicación, ejecute el siguiente comando:

```bash
go run cmd/lambda/main.go
```

# Pruebas

Para ejecutar las pruebas unitarias, ejecute el siguiente comando:

```bash
 go test -coverprofile=coverage.out ./internal/... && go tool cover -html=coverage.out -o coverage.html
```

# Despliegue De La Lambda

Para desplegar la lambda en AWS, ejecute el siguiente comando:

```bash
S=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap cmd/lambda/main.go  
zip function.zip bootstrap
aws lambda create-function \                       
  --function-name MyLambdaFunction \
  --handler bootstrap \
  --runtime provided.al2 \
  --role arn:aws:iam::xxxxxxxxxxxxxxxxx:role/LAMBDA_EXECUTION_ROLE \
  --zip-file fileb://function.zip \
  --environment Variables="{APP_ENV=dev,SECRET_NAME=gmf-secret,DB_HOST=gmfdb.cpqskyoqgdke.us-east-1.rds.amazonaws.com,DB_PORT=5432,DB_NAME=postgres,DB_USER=postgres,DB_PASSWORD=postgres,DB_SCHEMA=public,SMTP_SERVER=smtp.gmail.com,SMTP_PORT=587,SMTP_USER=juan.cantillo@sofka.com.co,SMTP_PASSWORD=nxqhdqdgflommpwg,SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/918665077918/MyQueue}"
```

## Creacion de la cola en AWS

Para crear la cola en AWS, ejecute el siguiente comando:

```bash
aws sqs create-queue --queue-name MyQueue
```

## envio de mensaje a la cola

Para enviar un mensaje a la cola, ejecute el siguiente comando:

```bash
aws sqs send-message --queue-url https://sqs.us-east-1.amazonaws.com/***********/MyQueue \
--message-body '{
  "id_plantilla": "PC001",
  "parametros": [
    {"nombre": "nombre_archivo", "valor": "TGMF-2024082801010001.txt"},
    {"nombre": "plataforma_origen", "valor": "STRATUS"},
    {"nombre": "fecha_recepcion", "valor": "07/10/2024"},
    {"nombre": "hora_recepcion", "valor": "09:19 AM"},
    {"nombre": "codigo_rechazo", "valor": "EPCM002"},
    {"nombre": "descripcion_rechazo", "valor": "Archivo ya existe con un estado no válido para su reproceso"}
  ]
}'
```






