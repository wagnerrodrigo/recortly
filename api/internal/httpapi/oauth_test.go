package httpapi

import (
	"testing"
	"time"

	"github.com/recortly/recortly/api/internal/config"
)

func testServer() *Server {
	return &Server{Cfg: config.Config{TokenKey: "chave-de-teste"}}
}

func TestStateRoundTrip(t *testing.T) {
	s := testServer()
	state := s.signState(42)
	uid, ok := s.verifyState(state)
	if !ok || uid != 42 {
		t.Fatalf("state deveria validar uid=42, veio %d ok=%v", uid, ok)
	}
}

func TestStateInvalido(t *testing.T) {
	s := testServer()
	tests := map[string]string{
		"vazio":        "",
		"sem ponto":    "abc",
		"assinatura":   "YWJj.ZGVm",
		"formato ruim": "YQ==.Yg==",
	}
	for name, state := range tests {
		t.Run(name, func(t *testing.T) {
			if _, ok := s.verifyState(state); ok {
				t.Fatalf("state %q não deveria validar", name)
			}
		})
	}
}

func TestStateExpirado(t *testing.T) {
	s := testServer()
	state := s.signStateWithExpiry(1, time.Now().Add(-time.Minute))
	if _, ok := s.verifyState(state); ok {
		t.Fatal("state expirado deveria ser rejeitado")
	}
}
