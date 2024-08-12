package oauth

type ProviderConfig struct {
	ID       string
	Secret   string
	Scope    string
	LoginURL string
	TokenURL string
	UserURL  string
}
