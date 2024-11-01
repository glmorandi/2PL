package main

import (
	"fmt"
	"sync"
)

// Operação de transação
type Operation struct {
	TxID int
	Type string
	Data string
}

// Transação
type Transaction struct {
	ID       int
	Active   bool
	Attempts int // Contador para tentativas de reinício
}

// Simulador do banco de dados
type Database struct {
	Data  map[string]int
	Locks map[string]int // 0 = nenhum lock, 1 = lock compartilhado, 2 = lock exclusivo
	mu    sync.Mutex
}

// Simulador do escalonador
type Scheduler struct {
	Database      *Database
	History       []Operation          // História de execução
	FinalHist     []Operation          // História final
	Delays        map[int][]Operation  // Mapeia TXID para operações em delay
	Transactions  map[int]*Transaction // Mapeia TXID para transações
	DeadlockCheck map[int]bool         // Para verificar deadlocks
}

// Inicializa o banco de dados
func NewDatabase() *Database {
	return &Database{
		Data:  map[string]int{"x": 1, "y": 2, "z": 3},
		Locks: make(map[string]int),
	}
}

// Inicializa o escalonador
func NewScheduler(db *Database) *Scheduler {
	return &Scheduler{
		Database:      db,
		History:       []Operation{},
		FinalHist:     []Operation{},
		Delays:        make(map[int][]Operation),
		Transactions:  make(map[int]*Transaction),
		DeadlockCheck: make(map[int]bool),
	}
}

// Concede lock para a operação
func (s *Scheduler) AcquireLock(op Operation) bool {
	s.Database.mu.Lock()
	defer s.Database.mu.Unlock()

	if op.Type == "r" {
		if s.Database.Locks[op.Data] == 0 || s.Database.Locks[op.Data] == 1 {
			s.Database.Locks[op.Data] = 1 // lock compartilhado
			fmt.Printf("Lock compartilhado concedido para T%d em %s\n", op.TxID, op.Data)
			return true
		}
	} else if op.Type == "w" {
		if s.Database.Locks[op.Data] == 0 {
			s.Database.Locks[op.Data] = 2 // lock exclusivo
			fmt.Printf("Lock exclusivo concedido para T%d em %s\n", op.TxID, op.Data)
			return true
		}
	}
	return false
}

// Libera lock para a operação
func (s *Scheduler) ReleaseLock(op Operation) {
	s.Database.mu.Lock()
	defer s.Database.mu.Unlock()

	if op.Type == "r" || op.Type == "w" {
		s.Database.Locks[op.Data] = 0
		fmt.Printf("Lock liberado para T%d em %s\n", op.TxID, op.Data)
	}
}

// Executa a operação
func (s *Scheduler) ExecuteOperation(op Operation) {
	s.History = append(s.History, op) // Adiciona operação à história
	if s.AcquireLock(op) {
		s.FinalHist = append(s.FinalHist, op)
		if op.Type == "r" {
			fmt.Printf("T%d: Lendo %s = %d\n", op.TxID, op.Data, s.Database.Data[op.Data])
		} else if op.Type == "w" {
			s.Database.Data[op.Data] += 1 // Exemplo de operação de escrita
			fmt.Printf("T%d: Escrevendo %s = %d\n", op.TxID, op.Data, s.Database.Data[op.Data])
		}
		s.ReleaseLock(op)
	} else {
		s.Delays[op.TxID] = append(s.Delays[op.TxID], op)
		fmt.Printf("T%d: Operação %s em delay\n", op.TxID, op.Type)
		s.DeadlockCheck[op.TxID] = true // Marca a transação como potencialmente em deadlock
		s.checkDeadlock(op.TxID)
	}
}

// Verifica deadlocks
func (s *Scheduler) checkDeadlock(txID int) {
	if s.DeadlockCheck[txID] {
		fmt.Printf("Deadlock detectado para T%d!\n", txID)
		s.abortTransaction(txID)
	}
}

// Aborta a transação e reinicia as operações
func (s *Scheduler) abortTransaction(txID int) {
	fmt.Printf("Abortando transação T%d...\n", txID)
	if ops, exists := s.Delays[txID]; exists {
		// Retira operações da história final
		for _, op := range ops {
			s.removeOperationFromFinalHist(op)
		}
		delete(s.Delays, txID)
	}
	// Reinicia a transação
	s.Transactions[txID].Attempts++
	s.restartTransaction(txID)
}

// Remove operação da história final
func (s *Scheduler) removeOperationFromFinalHist(op Operation) {
	for i, finalOp := range s.FinalHist {
		if finalOp.TxID == op.TxID && finalOp.Type == op.Type && finalOp.Data == op.Data {
			s.FinalHist = append(s.FinalHist[:i], s.FinalHist[i+1:]...)
			break
		}
	}
}

// Reinicia a transação
func (s *Scheduler) restartTransaction(txID int) {
	fmt.Printf("Reiniciando transação T%d, tentativas: %d\n", txID, s.Transactions[txID].Attempts)
	for _, op := range s.Delays[txID] {
		s.ExecuteOperation(op)
	}
}

// Processa a história de execução
func (s *Scheduler) ProcessHistory(operations []Operation) {
	var wg sync.WaitGroup
	for _, op := range operations {
		if _, exists := s.Transactions[op.TxID]; !exists {
			s.Transactions[op.TxID] = &Transaction{ID: op.TxID, Active: true}
		}
		wg.Add(1)
		go func(op Operation) {
			defer wg.Done()
			s.ExecuteOperation(op)
		}(op)
	}
	wg.Wait() // Aguarda todas as goroutines terminarem
	s.processDelays()
}

// Processa operações em delay
func (s *Scheduler) processDelays() {
	for txID, ops := range s.Delays {
		for _, op := range ops {
			if s.AcquireLock(op) {
				s.FinalHist = append(s.FinalHist, op)
				s.ReleaseLock(op)
				fmt.Printf("T%d: Operação %s executada após delay\n", op.TxID, op.Type)
				delete(s.Delays, txID)
				s.DeadlockCheck[txID] = false // Resetar checagem de deadlock
			}
		}
	}
}

func main() {
	db := NewDatabase()
	scheduler := NewScheduler(db)

	operations := []Operation{
		{TxID: 1, Type: "w", Data: "x"}, // T1 tenta escrever x
		{TxID: 2, Type: "w", Data: "y"}, // T2 tenta escrever y
		{TxID: 1, Type: "w", Data: "y"}, // T1 tenta escrever y
		{TxID: 2, Type: "w", Data: "x"}, // T2 tenta escrever x
	}

	scheduler.ProcessHistory(operations)

	fmt.Println("\nHistória de Execução (HI):")
	for _, op := range scheduler.History {
		fmt.Printf("T%d: %s %s\n", op.TxID, op.Type, op.Data)
	}

	fmt.Println("\nHistória Final de Execução (HF):")
	for _, op := range scheduler.FinalHist {
		fmt.Printf("T%d: %s %s\n", op.TxID, op.Type, op.Data)
	}
}
