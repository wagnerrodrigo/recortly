package store

import (
	"os"
	"path/filepath"
)

// absPath devolve caminho absoluto para arquivos salvos com caminho relativo.
// O worker pode rodar de um diretório diferente da API.
func absPath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	wd, err := os.Getwd()
	if err != nil {
		return p
	}
	return filepath.Join(wd, p)
}
