package httpapi

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/recortly/recortly/api/internal/model"
)

var invalidFileChars = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]+`)

// videoFileName monta o nome do arquivo a partir da frase da variação.
// Ex.: "02 - Olha so o que rolou nesse servidor.mp4"
func videoFileName(headline string, index, total int) string {
	clean := sanitizeName(headline)
	if clean == "" {
		clean = "video"
	}
	if total > 1 {
		return fmt.Sprintf("%02d - %s.mp4", index, clean)
	}
	return clean + ".mp4"
}

func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = invalidFileChars.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.Trim(s, " .-")
	// remove acentos para evitar problemas em players antigos
	s = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'â', 'ã', 'ä':
			return 'a'
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'ô', 'õ', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		case 'Á', 'À', 'Â', 'Ã', 'Ä':
			return 'A'
		case 'É', 'È', 'Ê', 'Ë':
			return 'E'
		case 'Í', 'Ì', 'Î', 'Ï':
			return 'I'
		case 'Ó', 'Ò', 'Ô', 'Õ', 'Ö':
			return 'O'
		case 'Ú', 'Ù', 'Û', 'Ü':
			return 'U'
		case 'Ç':
			return 'C'
		}
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) > 70 {
		s = string([]rune(s)[:70])
	}
	return strings.Trim(s, " .-")
}

// handleRenderZip baixa todas as variações de um render em um arquivo ZIP,
// nomeando cada vídeo com a headline usada.
func (s *Server) handleRenderZip(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	render, err := s.Store.GetRender(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if len(render.Outputs) == 0 {
		writeErr(w, http.StatusNotFound, "render sem arquivos")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="render-%d.zip"`, render.ID))
	zw := zip.NewWriter(w)
	defer zw.Close()

	for i, out := range render.Outputs {
		path := out.FilePath
		if path == "" {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		name := videoFileName(out.Headline, i+1, len(render.Outputs))
		entry, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			f.Close()
			return
		}
		_, copyErr := io.Copy(entry, f)
		f.Close()
		if copyErr != nil {
			return
		}
	}
}

var _ = model.RenderOutput{}
var _ = filepath.Base
