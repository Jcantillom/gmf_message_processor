package models

// SQSMessage representa la estructura esperada de un mensaje recibido desde SQS.
type SQSMessage struct {
	IDPlantilla string          `json:"IdPlantilla"`
	Parametro   []ParametrosSQS `json:"parametro"`
}

type ParametrosSQS struct {
	NombreArchivo      string `json:"nombre_archivo"`
	PlataformaOrigen   string `json:"plataforma_origen"`
	FechaRecepcion     string `json:"fecha_recepcion"`
	HoraRecepcion      string `json:"hora_recepcion"`
	CodigoRechazo      string `json:"codigo_rechazo"`
	DescripcionRechazo string `json:"descripcion_rechazo"`
	DetalleRechazo     string `json:"detalle_rechazo"`
}
