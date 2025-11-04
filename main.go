package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Event struct {
	ID        int               `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	RemoteIP  string            `json:"remote_ip"`
	CreatedAt time.Time         `json:"created_at"`
}

type Store struct {
	mu     sync.Mutex
	events []Event
	nextID int
	maxLen int
}

func NewStore(maxLen int) *Store {
	return &Store{
		events: make([]Event, 0),
		nextID: 1,
		maxLen: maxLen,
	}
}

func (s *Store) Add(e Event) Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	e.ID = s.nextID
	s.nextID++

	s.events = append(s.events, e)
	// mantém só os últimos maxLen
	if len(s.events) > s.maxLen {
		s.events = s.events[len(s.events)-s.maxLen:]
	}

	return e
}

func (s *Store) All() []Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Event, len(s.events))
	copy(out, s.events)
	return out
}

func (s *Store) GetByID(id int) (*Event, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID == id {
			evCopy := s.events[i]
			return &evCopy, true
		}
	}
	return nil, false
}

func main() {
	store := NewStore(100) // guarda até 100 eventos

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		hookHandler(w, r, store)
	})
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		eventsHandler(w, r, store)
	})
	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		eventByIDHandler(w, r, store)
	})

	addr := ":8082"
	log.Printf("Servidor rodando em http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	msg := `
Mini Webhook Inspector em Go

Endpoints disponíveis:

  POST /hook         -> envia qualquer requisição aqui (JSON, form, etc.)
  GET  /events       -> lista dos últimos eventos recebidos
  GET  /events/{id}  -> detalha um evento específico

Exemplo curl:

  curl -X POST http://localhost:8082/hook -H "Content-Type: application/json" -d '{"foo":"bar"}'
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(msg))
}

func hookHandler(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Use POST/PUT/PATCH para enviar eventos", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "erro ao ler body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for k, vals := range r.Header {
		headers[k] = strings.Join(vals, ", ")
	}

	ev := Event{
		Method:    r.Method,
		Path:      r.URL.Path,
		Headers:   headers,
		Body:      string(bodyBytes),
		RemoteIP:  r.RemoteAddr,
		CreatedAt: time.Now(),
	}

	ev = store.Add(ev)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ev)
}

func eventsHandler(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}

	events := store.All()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(events)
}

func eventByIDHandler(w http.ResponseWriter, r *http.Request, store *Store) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}

	// path esperado: /events/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/events/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "ID não informado", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	ev, ok := store.GetByID(id)
	if !ok {
		http.Error(w, "Evento não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(ev)
}
