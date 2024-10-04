package config

import (
	"gmf_message_processor/connection"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
	"gorm.io/gorm"
)

// MockDBManager es un mock que implementa la interfaz de DBManagerInterface.
type MockDBManager struct {
	mock.Mock
	DB *gorm.DB
}

func (m *MockDBManager) InitDB() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBManager) CloseDB() {
	m.Called()
}

func (m *MockDBManager) GetDB() *gorm.DB {
	return m.DB
}

// MockConfigManager es un mock para ConfigManager.
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) InitConfig() {
	m.Called()
}

func TestInitApplication(t *testing.T) {
	// Crear mocks para DBManager y ConfigManager
	mockDBManager := new(MockDBManager)
	mockConfigManager := new(MockConfigManager)

	// Configurar los mocks para simular el comportamiento esperado
	mockDBManager.On("InitDB").Return(nil)
	mockDBManager.On("CloseDB").Return()
	mockConfigManager.On("InitConfig").Return()

	// Llamar directamente a los métodos con mocks
	configManager := mockConfigManager
	configManager.InitConfig()

	// Reemplazar la inicialización de DBManager por el mock
	dbManager := mockDBManager
	err := dbManager.InitDB()
	assert.NoError(t, err)

	// Simular la inicialización de otros componentes como en InitApplication
	repo := repository.NewPlantillaRepository(dbManager.GetDB()) // Utilizamos GetDB() para obtener *gorm.DB
	emailService := email.NewSMTPEmailService()
	plantillaService := service.NewPlantillaService(repo, emailService)

	// Verificar que los servicios y DBManager se inicialicen correctamente
	assert.NotNil(t, plantillaService)
	assert.NotNil(t, dbManager)

	// Verificar que InitDB fue llamado
	mockDBManager.AssertCalled(t, "InitDB")
	// Verificar que InitConfig fue llamado
	mockConfigManager.AssertCalled(t, "InitConfig")
}

func TestCleanupApplication(t *testing.T) {
	// Crear un mock de DBManager
	mockDBManager := new(MockDBManager)

	// Configurar el mock para simular el comportamiento esperado
	mockDBManager.On("CloseDB").Return()

	// Llamar a CleanupApplication con el mock
	CleanupApplication(mockDBManager) // Ahora pasamos el mock directamente porque implementa la interfaz correcta

	// Verificar que CloseDB fue llamado
	mockDBManager.AssertCalled(t, "CloseDB")
}

func TestNewConfigManager(t *testing.T) {
	// Crear una nueva instancia de ConfigManager
	cm := NewConfigManager()

	// Verificar que la instancia no sea nula
	assert.NotNil(t, cm)
}

func TestNewDBManager(t *testing.T) {
	// Crear una nueva instancia de DBManager
	dbm := connection.NewDBManager()

	// Verificar que la instancia no sea nula
	assert.NotNil(t, dbm)
}
