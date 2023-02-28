package verifier

import (
	"net/http"

	"github.com/slack-go/slack"
)

type Verifier struct {
	signingSecret string
}

func New(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret: signingSecret,
	}
}

func (v *Verifier) Verify(header http.Header, body []byte) error {
	sv, err := slack.NewSecretsVerifier(header, v.signingSecret)
	if err != nil {
		return err
	}

	_, err = sv.Write(body)
	if err != nil {
		return err
	}

	err = sv.Ensure()
	if err != nil {
		return err
	}

	return nil
}
