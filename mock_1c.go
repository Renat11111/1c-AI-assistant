package main

import (
	"fmt"
	"strings"
)

// MockDB представляет собой имитацию базы данных 1С.
type MockDB struct {
	products map[string]int
	debts    map[string]float64
}

// NewMockDB создает и инициализирует новую имитацию БД.
func NewMockDB() *MockDB {
	return &MockDB{
		products: map[string]int{
			"стол офисный модель а": 152,
			"стул офисный стандарт": 312,
			"монитор 24 дюйма":      88,
			"клавиатура беспроводная": 210,
		},
		debts: map[string]float64{
			"ооо ромашка": 125430.50,
			"ооо лютик":   0,
			"ип васильев": 30000.00,
		},
	}
}

// GetStockBalance имитирует запрос остатка товара на складе.
// Для упрощения, поиск нечеткий (игнорирует регистр).
func (db *MockDB) GetStockBalance(productName string) (int, error) {
	productName = strings.ToLower(productName)
	balance, ok := db.products[productName]
	if !ok {
		return 0, fmt.Errorf("товар '%s' не найден в базе данных", productName)
	}
	return balance, nil
}

// GetCounterpartyDebt имитирует запрос задолженности контрагента.
func (db *MockDB) GetCounterpartyDebt(counterpartyName string) (float64, error) {
	counterpartyName = strings.ToLower(counterpartyName)
	debt, ok := db.debts[counterpartyName]
	if !ok {
		return 0, fmt.Errorf("контрагент '%s' не найден в базе данных", counterpartyName)
	}
	return debt, nil
}
