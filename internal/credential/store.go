package credential

type Store interface {
	Set(agent, serverURL, secret string) error
	Get(agent, serverURL string) (string, error)
	Delete(agent, serverURL string) error
}
