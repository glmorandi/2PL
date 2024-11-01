#### Características do Funcionamento

1. **Controle de Locks**:
   - O sistema implementa dois tipos de locks: compartilhados (para leitura) e exclusivos (para escrita).
   - Uma operação de leitura (r) pode ser realizada se não houver um lock exclusivo ativo. Um lock compartilhado é concedido se já houver um lock compartilhado ou se não houver locks.
   - Uma operação de escrita (w) requer um lock exclusivo, que é concedido se não houver locks ativos.

2. **Detecção e Tratamento de Deadlocks**:
   - O sistema utiliza um mecanismo simples para detectar deadlocks. Se uma transação não consegue obter um lock e entra em delay, é marcada como potencialmente em deadlock.
   - Se um deadlock é detectado, a transação problemática é abortada, e suas operações em delay são removidas da história final. A transação é reiniciada, e as operações são tentadas novamente.

3. **Reinício de Transações**:
   - Após um abort, a transação é reiniciada e suas operações que estavam em delay são processadas.

#### Características Implementadas e Não Implementadas

- **Implementadas**:
  - Gerenciamento de locks (compartilhados e exclusivos).
  - Detecção de deadlocks e abortos de transações.
  - Processamento de operações em delay.
  - Manutenção de uma história de execução e uma história final.

- **Não Implementadas**:
  - Timeout para operações em delay; transações podem ficar em delay indefinidamente.
  - Sistema de prioridade entre transações, o que poderia ser usado para decidir quais transações abortar em caso de deadlock.
  
#### Entrada dos Dados

Os dados de entrada consistem em uma lista de operações, cada uma representando uma transação que pode ser uma operação de leitura ("r") ou escrita ("w") em um dado específico (ex: "x", "y", "z"). Cada operação é identificada por um ID de transação (`TxID`).

Exemplo de entrada:
```go
operations := []Operation{
    {TxID: 1, Type: "w", Data: "x"}, // T1 tenta escrever x
    {TxID: 2, Type: "w", Data: "y"}, // T2 tenta escrever y
    {TxID: 1, Type: "w", Data: "y"}, // T1 tenta escrever y
    {TxID: 2, Type: "w", Data: "x"}, // T2 tenta escrever x
}
```

#### Saída de Dados

O sistema gera duas saídas principais:
1. **História de Execução (HI)**: Um registro de todas as operações executadas, incluindo aquelas que entraram em delay.
2. **História Final de Execução (HF)**: Um registro das operações que foram completadas com sucesso, após resolução de delays e deadlocks.

A saída é impressa no console, mostrando o status de cada operação, incluindo mensagens sobre concessão e liberação de locks, além de qualquer detecção de deadlock e reinício de transações.

#### Estruturas de Dados

1. **Operação (`Operation`)**:
```go
type Operation struct {
	TxID int
	Type string
	Data string
}
```

2. **Transação (`Transaction`)**:
```go
type Transaction struct {
	ID       int
	Active   bool
	Attempts int
}
```

3. **Banco de Dados (`Database`)**:
```go
type Database struct {
	Data  map[string]int
	Locks map[string]int
	mu    sync.Mutex
}
```
4. **Escalonador (`Scheduler`)**: 
```go
type Scheduler struct {
	Database      *Database
	History       []Operation
	FinalHist     []Operation          
	Delays        map[int][]Operation
	Transactions  map[int]*Transaction 
	DeadlockCheck map[int]bool         
}
```