package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/server/restapi/services" // Для NewSingleAgentLoader
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool" // Для tool.New
	"google.golang.org/genai"
)

// Структура для аргументов инструмента "get_stock_balance"
type GetStockBalanceParams struct {
	ProductName string `json:"product_name"`
}

// Структура для аргументов инструмента "get_counterparty_debt"
type GetCounterpartyDebtParams struct {
	CounterpartyName string `json:"counterparty_name"`
}

// Структуры для результатов инструментов.
type GetStockBalanceOutput struct {
	StockBalance int `json:"stock_balance"`
}
type GetCounterpartyDebtOutput struct {
	Debt float64 `json:"debt"`
}

func main() {
	ctx := context.Background()

	// Проверяем наличие API-ключа.
	// Агент не будет работать без него.
	if os.Getenv("GOOGLE_API_KEY") == "" {
		log.Fatal("Ошибка: Переменная окружения GOOGLE_API_KEY не установлена. Пожалуйста, установите ваш API-ключ.")
	}

	// Инициализируем нашу имитацию базы данных 1С.
	db := NewMockDB()

	// Определяем инструменты, которые будут доступны AI-агенту.
	// Каждый инструмент имеет имя, описание (чтобы AI понял, когда его использовать)
	// и функцию, которую нужно выполнить.
	tools := []tool.Tool{
		func() tool.Tool {
			handler := func(ctx tool.Context, input GetStockBalanceParams) GetStockBalanceOutput {
				balance, err := db.GetStockBalance(input.ProductName)
				if err != nil {
					log.Printf("Ошибка в инструменте get_stock_balance: %v", err)
					return GetStockBalanceOutput{}
				}
				return GetStockBalanceOutput{StockBalance: balance}
			}
			t, err := functiontool.New(functiontool.Config{
				Name:        "get_stock_balance",
				Description: "Получить остаток товара на складе по его названию.",
			}, handler)
			if err != nil {
				log.Fatalf("Failed to create tool get_stock_balance: %v", err)
			}
			return t
		}(),
		func() tool.Tool {
			handler := func(ctx tool.Context, input GetCounterpartyDebtParams) GetCounterpartyDebtOutput {
				debt, err := db.GetCounterpartyDebt(input.CounterpartyName)
				if err != nil {
					log.Printf("Ошибка в инструменте get_counterparty_debt: %v", err)
					return GetCounterpartyDebtOutput{}
				}
				return GetCounterpartyDebtOutput{Debt: debt}
			}
			t, err := functiontool.New(functiontool.Config{
				Name:        "get_counterparty_debt",
				Description: "Получить текущую задолженность клиента (контрагента) по его названию.",
			}, handler)
			if err != nil {
				log.Fatalf("Failed to create tool get_counterparty_debt: %v", err)
			}
			return t
		}(),
	}

	// Инициализируем модель Gemini.
	// API-ключ считывается из переменной окружения GOOGLE_API_KEY.
	model, err := gemini.NewModel(ctx, "gemini-2.0-flash-001", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Ошибка при создании модели Gemini: %v", err)
	}

	// Создаем нашего агента.
	agent, err := llmagent.New(llmagent.Config{
		Name:        "1c_assistant", // Имя для агента
		Model:       model,
		Instruction: "Ты — дружелюбный и эффективный AI-ассистент для работы с системой 1С. Твоя задача — отвечать на вопросы пользователя, используя предоставленные тебе инструменты. Всегда отвечай на русском языке.",
		Tools:       tools,
	})
	if err != nil {
		log.Fatalf("Ошибка при создании LLM агента: %v", err)
	}

	// Запускаем агент с помощью full launcher.
	// Это стандартный способ запуска, который предоставляет CLI,
	// веб-интерфейс и другие способы взаимодействия.
	config := &launcher.Config{
		AgentLoader: services.NewSingleAgentLoader(agent),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
