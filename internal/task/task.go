package task

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// TaskStatus обозначает статус задачи.
type TaskStatus string

const (
	Pending   TaskStatus = "pending"
	Completed TaskStatus = "completed"
)

// Task представляет собой вычислительную операцию, получаемую агентом.
type Task struct {
	ID            int        `json:"id"`
	ExpressionID  int        `json:"-"`
	Arg1          float64    `json:"arg1"`
	Arg2          float64    `json:"arg2"`
	Operation     string     `json:"operation"`      // "addition", "subtraction", "multiplication", "division"
	OperationTime int        `json:"operation_time"` // время выполнения в миллисекундах
	Result        float64    `json:"result"`
	Status        TaskStatus `json:"status"`
}

// Node представляет собой узел дерева арифметического выражения.
type Node struct {
	IsOperator bool
	Operator   string  // "+", "-", "*", "/"
	Value      float64 // для чисел и вычисленных операторов
	Left       *Node
	Right      *Node
	Parent     *Node
	TaskID     int   // номер задачи, если узел относится к операции
	Computed   bool  // true, если значение вычислено (либо для литерала, либо после выполнения операции)
}

// Expression содержит исходное выражение, его дерево и список задач для вычисления.
type ExpressionStatus string

const (
	ExprPending   ExpressionStatus = "pending"
	ExprCompleted ExpressionStatus = "completed"
)

type Expression struct {
	ID     int
	Raw    string
	Root   *Node
	Tasks  []*Task
	Status ExpressionStatus
	Result float64
}

// --- Парсер арифметического выражения ---

// Реализуем простой рекурсивный спуск для арифметических выражений.
// Грамматика:
//   expr   := term { ('+'|'-') term }
//   term   := factor { ('*'|'/') factor }
//   factor := number | '(' expr ')'

type parser struct {
	input string
	pos   int
}

func (p *parser) current() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) consume() {
	p.pos++
}

func (p *parser) skipSpaces() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *parser) parseNumber() (*Node, error) {
	start := p.pos
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return nil, errors.New("expected number")
	}
	numStr := p.input[start:p.pos]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, err
	}
	// Числовой узел считается вычисленным
	return &Node{IsOperator: false, Value: val, Computed: true}, nil
}

func (p *parser) parseFactor() (*Node, error) {
	p.skipSpaces()
	if p.current() == '(' {
		p.consume() // '('
		node, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.skipSpaces()
		if p.current() != ')' {
			return nil, errors.New("expected ')'")
		}
		p.consume() // ')'
		return node, nil
	}
	return p.parseNumber()
}

func (p *parser) parseTerm() (*Node, error) {
	node, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	for {
		p.skipSpaces()
		op := p.current()
		if op == '*' || op == '/' {
			p.consume()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			newNode := &Node{
				IsOperator: true,
				Operator:   string(op),
				Left:       node,
				Right:      right,
				Computed:   false,
			}
			node.Parent = newNode
			right.Parent = newNode
			node = newNode
		} else {
			break
		}
		p.skipSpaces()
	}
	return node, nil
}

func (p *parser) parseExpr() (*Node, error) {
	node, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	for {
		p.skipSpaces()
		op := p.current()
		if op == '+' || op == '-' {
			p.consume()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			newNode := &Node{
				IsOperator: true,
				Operator:
				Operator:   string(op),
				Left:       node,
				Right:      right,
				Computed:   false,
			}
			node.Parent = newNode
			right.Parent = newNode
			node = newNode
		} else {
			break
		}
		p.skipSpaces()
	}
	return node, nil
}

// ParseExpression parses the input string into an expression tree.
func ParseExpression(input string) (*Expression, error) {
	p := &parser{input: strings.TrimSpace(input)}
	root, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.input) {
		return nil, errors.New("unexpected characters at the end of input")
	}
	return &Expression{
		Raw:    input,
		Root:   root,
		Status: ExprPending,
	}, nil
}

// GetOperationTime returns the time required to perform an operation based on environment variables.
func GetOperationTime(operation string) (int, error) {
	var envVar string
	switch operation {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 0, errors.New("unsupported operation")
	}
	timeStr := os.Getenv(envVar)
	if timeStr == "" {
		return 0, errors.New("environment variable not set: " + envVar)
	}
	timeMs, err := strconv.Atoi(timeStr)
	if err != nil {
		return 0, errors.New("invalid value for environment variable: " + envVar)
	}
	return timeMs, nil
}

// CreateTasksFromExpression traverses the expression tree and creates tasks for each operation.
func CreateTasksFromExpression(expr *Expression, exprID int) ([]*Task, error) {
	var tasks []*Task
	taskID := 1

	var traverse func(node *Node) error
	traverse = func(node *Node) error {
		if node == nil {
			return nil
		}

		// Traverse left and right subtrees
		if err := traverse(node.Left); err != nil {
			return err
		}
		if err := traverse(node.Right); err != nil {
			return err
		}

		// If the node is an operator, create a task for it
		if node.IsOperator {
			operationTime, err := GetOperationTime(node.Operator)
			if err != nil {
				return err
			}

			task := &Task{
				ID:            taskID,
				ExpressionID:  exprID,
				Operation:     node.Operator,
				OperationTime: operationTime,
				Status:        Pending,
			}

			// Link arguments to the task
			if node.Left != nil && node.Left.Computed {
				task.Arg1 = node.Left.Value
			} else if node.Left != nil {
				task.Arg1TaskID = node.Left.TaskID
			}

			if node.Right != nil && node.Right.Computed {
				task.Arg2 = node.Right.Value
			} else if node.Right != nil {
				task.Arg2TaskID = node.Right.TaskID
			}

			tasks = append(tasks, task)
			node.TaskID = taskID
			taskID++
		}

		return nil
	}

	if err := traverse(expr.Root); err != nil {
		return nil, err
	}

	expr.Tasks = tasks
	return tasks, nil
}

// EvaluateTask computes the result of a task based on its operation and arguments.
func EvaluateTask(task *Task, tasks map[int]*Task) (float64, error) {
	var arg1, arg2 float64

	// Resolve Arg1
	if task.Arg1TaskID != 0 {
		arg1Task, exists := tasks[task.Arg1TaskID]
		if !exists || arg1Task.Status != Completed {
			return 0, errors.New("dependent task not completed")
		}
		arg1 = arg1Task.Result
	} else {
		arg1 = task.Arg1
	}

	// Resolve Arg2
	if task.Arg2TaskID != 0 {
		arg2Task, exists := tasks[task.Arg2TaskID]
		if !exists || arg2Task.Status != Completed {
			return 0, errors.New("dependent task not completed")
		}
		arg2 = arg2Task.Result
	} else {
		arg2 = task.Arg2
	}

	// Perform the operation
	switch task.Operation {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, errors.New("division by zero")
		}
		return arg1 / arg2, nil
	default:
		return 0, errors.New("unsupported operation")
	}
}