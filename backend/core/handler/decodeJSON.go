package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// decodeJSON aceita exatamente um objeto JSON e rejeita campos desconhecidos
// ou conteúdo extra após o primeiro valor.
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("mais de um valor JSON no body")
		}
		return err
	}

	return nil
}
