package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"calc_service/internal/task"

	"github.com/gorilla/mux"
)

type Server struct {
	Router       *mux.Router
	expressions  map[int]*task.Expression
	tasks        []*task.Task             
	taskMutex    sync.Mutex               
	exprMutex    sync.Mutex               
	nextExprID   int                      
	nextTaskID   int                     
}

func NewServer() *Server {
	s := &Server{
		Router:      mux.NewRouter(),
		expressions: make(map[int]*task.Expression),
	}

	s.Router.HandleFunc("/api/v1/calculate", s.handleCalculate).Methods("POST")
	s.Router.HandleFunc("/api/v1/expressions", s.handleGetExpressions).Methods("GET")
	s.Router.HandleFunc("/api/v1/expressions/{id:[0-9]+}", s.handleGetExpressionByID).Methods("GET")
	s.Router.HandleFunc("/internal/task", s.handleGetTask).Methods("GET")
	s.Router.HandleFunc("/internal/task", s.handlePostTaskResult).Methods("POST")

	return s
}

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Expression == "" {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	s.exprMutex.Lock()
	exprID := s.nextExprID
	s.nextExprID++
	expr := task.NewExpression(exprID, req.Expression)
	s.expressions[exprID] = expr
	s.exprMutex.Unlock()

	tasks, err := expr.ParseToTasks(s.nextTaskID)
	if err != nil {
		http.Error(w, "Failed to parse expression", http.StatusUnprocessableEntity)
		return
	}

	s.taskMutex.Lock()
	for _, t := range tasks {
		t.ID = s.nextTaskID
		s.nextTaskID++
		s.tasks = append(s.tasks, t)
	}
	s.taskMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": exprID})
}

func (s *Server) handleGetExpressions(w http.ResponseWriter, r *http.Request) {
	s.exprMutex.Lock()
	defer s.exprMutex.Unlock()

	var expressions []map[string]interface{}
	for _, expr := range s.expressions {
		expressions = append(expressions, map[string]interface{}{
			"id":     expr.ID,
			"status": expr.Status,
			"result": expr.Result,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (s *Server) handleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	s.exprMutex.Lock()
	defer s.exprMutex.Unlock()

	expr, exists := s.expressions[id]
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"expression": map[string]interface{}{
			"id":     expr.ID,
			"status": expr.Status,
			"result": expr.Result,
		},
	})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if len(s.tasks) == 0 {
		http.Error(w, "No tasks available", http.StatusNotFound)
		return
	}

	task := s.tasks[0]
	s.tasks = s.tasks[1:]

	json.NewEncoder(w).Encode(map[string]*task.Task{"task": task})
}

func (s *Server) handlePostTaskResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID < 0 {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	s.exprMutex.Lock()
	defer s.exprMutex.Unlock()

	for _, expr := range s.expressions {
		for _, t := range expr.Tasks {
			if t.ID == req.ID && t.Status == task.Pending {
				t.Result = req.Result
				t.Status = task.Completed

				if expr.CheckIfCompleted() {
					expr.CalculateFinalResult()
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}
