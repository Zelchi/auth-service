package handler

import (
	"authentication/core/observability"
	"fmt"
	"net/http"
)

func Metrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprintf(w, "# HELP auth_verification_email_failures_total Falhas ao enviar códigos de verificação.\n")
	fmt.Fprintf(w, "# TYPE auth_verification_email_failures_total counter\n")
	fmt.Fprintf(w, "auth_verification_email_failures_total %d\n", observability.VerificationEmailFailures())
}
