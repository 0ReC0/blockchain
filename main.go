package main

import (
	"crypto/sha256" // Пакет для хэширования данных с использованием SHA-256
	"encoding/hex"  // Для кодирования байтов в шестнадцатеричную строку
	"fmt"           // Для форматированного ввода-вывода
	"strconv"       // Для преобразования между строками и числами
	"strings"       // Для работы с строками
	"time"          // Для работы со временем
)

// Структура блока в блокчейне
type Block struct {
	Index     int    // Порядковый номер блока
	Timestamp string // Время создания блока
	Data      string // Данные, хранящиеся в блоке
	PrevHash  string // Хэш предыдущего блока (для связи блоков)
	Hash      string // Хэш текущего блока
	Nonce     int    // Число, используемое для майнинга (Proof of Work)
}

var blockchain []Block // Глобальная переменная для хранения всей цепочки блоков

// Функция для вычисления SHA-256 хэша блока
func calculateHash(block Block) string {
	// Объединяем все поля блока в одну строку
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Data + block.PrevHash + strconv.Itoa(block.Nonce)
	h := sha256.New()
	h.Write([]byte(record)) // Хэшируем объединенную строку
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed) // Возвращаем хэш в виде шестнадцатеричной строки
}

// Функция для генерации нового блока
func generateBlock(oldBlock Block, data string) Block {
	var newBlock Block
	newBlock.Index = oldBlock.Index + 1      // Индекс нового блока на 1 больше предыдущего
	newBlock.Timestamp = time.Now().String() // Устанавливаем текущее время
	newBlock.Data = data                     // Устанавливаем данные блока
	newBlock.PrevHash = oldBlock.Hash        // Хэш предыдущего блока
	newBlock.Nonce = 0                       // Начальное значение Nonce
	newBlock.Hash = calculateHash(newBlock)  // Вычисляем хэш нового блока
	return newBlock
}

// Проверка валидности блока
func isBlockValid(newBlock, oldBlock Block) bool {
	// Проверяем, что индекс нового блока на 1 больше индекса предыдущего
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	// Проверяем, что хэш предыдущего блока совпадает с PrevHash нового блока
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	// Проверяем, что вычисленный хэш нового блока совпадает с его полем Hash
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

// // Замена текущей цепочки на новую (если новая цепочка длиннее)
// func replaceChain(newBlocks []Block) {
// 	if len(newBlocks) > len(blockchain) {
// 		blockchain = newBlocks // Заменяем цепочку
// 	}
// }

// Реализация Proof of Work (майнинг): ищем хэш, начинающийся с "0000"
func proofOfWork(block Block) Block {
	for {
		hash := calculateHash(block)
		if strings.HasPrefix(hash, "0000") { // Условие: хэш начинается с 4 нулей
			block.Hash = hash // Сохраняем подходящий хэш
			return block
		}
		block.Nonce++ // Увеличиваем Nonce и пробуем снова
	}
}

func main() {
	// Создаем генезис-блок (первый блок цепочки)
	genesisBlock := Block{0, time.Now().String(), "Genesis Block", "", "", 0}
	genesisBlock.Hash = calculateHash(genesisBlock) // Вычисляем хэш генезис-блока
	blockchain = append(blockchain, genesisBlock)   // Добавляем его в цепочку

	// Генерируем 9 новых блоков
	for i := 1; i < 10; i++ {
		oldBlock := blockchain[len(blockchain)-1]                       // Берем последний блок из цепочки
		newBlock := generateBlock(oldBlock, fmt.Sprintf("Block %d", i)) // Создаем новый блок
		newBlock = proofOfWork(newBlock)                                // Применяем Proof of Work для майнинга
		if isBlockValid(newBlock, oldBlock) {                           // Проверяем валидность нового блока
			blockchain = append(blockchain, newBlock) // Добавляем новый блок в цепочку
		}
	}

	// Выводим информацию о всех блоках в цепочке
	for _, block := range blockchain {
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %s\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Previous Hash: %s\n", block.PrevHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println()
	}
}
