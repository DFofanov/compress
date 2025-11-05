package main

import (
	"log"

	"compressor/internal/domain/entities"
	"compressor/internal/domain/repositories"
	"compressor/internal/infrastructure/compressors"
	"compressor/internal/infrastructure/config"
	"compressor/internal/infrastructure/logging"
	infraRepos "compressor/internal/infrastructure/repositories"
	"compressor/internal/presentation/tui"
	usecases "compressor/internal/usecase"
)

func main() {
	// Загрузка конфигурации
	configRepo := config.NewRepository()
	appConfig, err := configRepo.Load("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация базового логгера (в файл)
	fileLogger, err := logging.NewFileLogger(
		appConfig.Output.LogFileName,
		appConfig.Output.LogLevel,
		appConfig.Output.LogMaxSizeMB,
		appConfig.Output.LogToFile,
	)
	if err != nil {
		log.Printf("Предупреждение: не удалось инициализировать логгер: %v", err)
	}
	if fileLogger != nil {
		defer fileLogger.Close()
	}

	// Инициализация TUI
	tuiManager := tui.NewManager()
	tuiManager.Initialize()

	// Оборачиваем логгер адаптером, чтобы видеть логи в TUI
	var logger repositories.Logger
	logger = tui.NewUILogger(fileLogger, tuiManager)

	// Инициализация репозиториев
	fileRepo := infraRepos.NewFileSystemRepository()
	compressionConfigRepo := infraRepos.NewConfigRepository()

	// Выбираем компрессор на основе конфигурации
	var compressor repositories.PDFCompressor
	switch appConfig.Compression.Algorithm {
	case "unipdf":
		compressor = compressors.NewUniPDFCompressor()
	default:
		compressor = compressors.NewPDFCPUCompressor()
	}

	// Инициализация компрессора изображений
	imageCompressor := compressors.NewImageCompressor()

	// Инициализация use cases
	processUseCase := usecases.NewProcessPDFsUseCase(
		compressor,
		fileRepo,
		compressionConfigRepo,
		logger,
	)

	imageUseCase := usecases.NewCompressImageUseCase(logger, imageCompressor)

	// Создаем объединенный процессор для всех типов файлов
	allFilesUseCase := usecases.NewProcessAllFilesUseCase(processUseCase, imageUseCase, logger)

	// Подключаем репортер прогресса к TUI
	processUseCase.SetProgressReporter(func(s entities.ProcessingStatus) {
		tuiManager.SendStatusUpdate(s)
	})

	// Создание процессора для обработки команд
	processor := NewApplicationProcessor(
		processUseCase,
		allFilesUseCase,
		appConfig,
		tuiManager,
		logger,
	)
	defer processor.Shutdown()

	// Привязываем запуск обработки к TUI
	tuiManager.SetOnStartProcessing(func() {
		// Получаем актуальную конфигурацию из TUI
		processor.config = tuiManager.GetConfig()
		go processor.StartProcessing()
	})

	// Автозапуск, если включен в конфигурации
	if appConfig.Compression.AutoStart {
		go processor.StartProcessing()
	}

	// Запуск TUI
	if err := tuiManager.Run(); err != nil {
		log.Fatalf("Ошибка запуска TUI: %v", err)
	}

	// Cleanup при выходе
	tuiManager.Cleanup()
}
