package logs

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
)

var Log = InitLogger()

// InitLogger inicializa el logger de logrus.
func InitLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

func LogError(ctx context.Context, msg string, args ...interface{}) {
	Log.WithContext(ctx).Errorf(msg, args...)
}

func LogInfo(ctx context.Context, msg string, args ...interface{}) {
	Log.WithContext(ctx).Infof(msg, args...)
}

func LogWarn(ctx context.Context, msg string, args ...interface{}) {
	Log.WithContext(ctx).Warnf(msg, args...)
}

func LogProcesandoMensajeSQS(ctx context.Context) {
	LogInfo(ctx, "Procesando mensaje de SQS... 🚀")
}

func LogMensajeProcesadoConExito(ctx context.Context) {
	Log.Info("Mensaje procesado con éxito ✅")
}

func LogPlantillaNoEncontrada(ctx context.Context, idPlantilla string) {
	LogError(ctx, "La plantilla con ID %s no existe en la base de datos  ❌ ", idPlantilla)
}

func LogPlantillaEncontrada(ctx context.Context, idPlantilla string) {
	LogInfo(
		ctx, "Plantilla con ID %s encontrada. Procediendo a enviar el correo electrónico... ✉️", idPlantilla)
}

func LogCorreoEnviado(ctx context.Context, idPlantilla string) {
	LogInfo(ctx, "Correo electrónico enviado exitosamente para IDPlantilla: %s ✅", idPlantilla)
}

func LogErrorEnvioCorreo(ctx context.Context, idPlantilla string, err error) {
	LogError(ctx, "Error enviando el correo para la plantilla con ID %s: %v ❌", idPlantilla, err)
}

func LogParametrosNoProporcionados(ctx context.Context, idPlantilla string) {
	LogError(ctx, "No se proporcionaron parámetros para la plantilla con ID %s", idPlantilla)
}

func LogFormatoMensajeValido(ctx context.Context) {
	LogInfo(ctx, "Formato de mensaje válido 😉")
}

func LogPlantillaInsertada(ctx context.Context, idPlantilla string) {
	LogInfo(ctx, "Plantilla con ID %s insertada correctamente en la base de datos 🌱", idPlantilla)
}

func LogDatosSemillaPlantillaInsertados(ctx context.Context) {
	LogInfo(ctx, "Datos de semilla de plantilla insertados correctamente en la base de datos  🍁")
}

func LogEnviandoCorreosADestinatarios(ctx context.Context, destinatarios string, toJSON []byte) {
	LogInfo(ctx, "Enviando correo electrónico a ... : %s 📤\n%s", destinatarios, toJSON)
}

func LogCorreosEnviados(ctx context.Context, destinatarios string) {
	LogInfo(ctx, "Correo electrónico enviado con éxito a  ✅ :\n%s", destinatarios)
}

func LogConexionBaseDatosEstablecida() {
	LogInfo(context.Background(), "Conexión a la base de datos establecida correctamente 🐘")
}

func LogErrorConexionBaseDatos(err error) {
	LogError(context.Background(), "Error al establecer la conexión a la base de datos: %v", err)
}

func LogErrorMigracionTablaPlantilla(err error) error {
	LogError(context.Background(), "Error al migrar la tabla Plantilla: %v", err) // <--- Agregado 'err' como argumento
	return err
}

func LogErrorCerrandoConexionBaseDatos(err error) {
	LogError(context.Background(), "Error al cerrar la conexión a la base de datos: %v", err)
}

func LogConexionBaseDatosCerrada() {
	LogInfo(context.Background(), "Conexión a la base de datos cerrada correctamente 🚪")
}

func LogMigracionTablaPlantillaCompletada() {
	LogInfo(context.Background(), "Migración de la tabla Plantilla completada. 🚀")
}

func LogArchivoEnvNoEncontrado() {
	LogInfo(context.Background(), "No se encontró el archivo .env, confiando en las variables de entorno.")
}
