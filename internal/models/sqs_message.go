package models

// SQSMessage representa la estructura esperada de un mensaje recibido desde SQS.
type SQSMessage struct {
	IDPlantilla string          `json:"id_plantilla"`
	Parametro   []ParametrosSQS `json:"parametros"`
	RetryCount  int             `json:"retry_count"`
}

type ParametrosSQS struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
}
