package main

import (
	"compressor/internal/domain/entities"
	"compressor/internal/domain/repositories"
	"compressor/internal/presentation/tui"
	usecases "compressor/internal/usecase"
	"context"
	"sync"
)

// ApplicationProcessor обрабатывает команды приложения
type ApplicationProcessor struct {
	processUseCase  *usecases.ProcessPDFsUseCase
	allFilesUseCase *usecases.ProcessAllFilesUseCase
	config          *entities.Config
	tuiManager      *tui.Manager
	logger          repositories.Logger

	// Graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewApplicationProcessor создает новый процессор приложения
func NewApplicationProcessor(
	processUseCase *usecases.ProcessPDFsUseCase,
	allFilesUseCase *usecases.ProcessAllFilesUseCase,
	config *entities.Config,
	tuiManager *tui.Manager,
	logger repositories.Logger,
) *ApplicationProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ApplicationProcessor{
		processUseCase:  processUseCase,
		allFilesUseCase: allFilesUseCase,
		config:          config,
		tuiManager:      tuiManager,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// StartProcessing запускает обработку всех поддерживаемых файлов
func (p *ApplicationProcessor) StartProcessing() {
	p.wg.Add(1)
	defer p.wg.Done()

	if p.logger != nil {
		supportedTypes := p.allFilesUseCase.GetSupportedFileTypes(p.config)
		p.logger.Info("Запуск обработки файлов. Поддерживаемые типы: %v", supportedTypes)
	}

	// Запускаем обработку всех поддерживаемых файлов
	if err := p.allFilesUseCase.Execute(p.config); err != nil {
		if p.logger != nil {
			p.logger.Error("Ошибка обработки: %v", err)
		}
		return
	}

	if p.logger != nil {
		p.logger.Success("Обработка файлов завершена успешно")
	}
}

// Shutdown корректно завершает работу процессора
func (p *ApplicationProcessor) Shutdown() {
	p.cancel()
	p.wg.Wait()
}
