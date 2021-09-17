package shortlink

type KeyType string

const (
	KeyTypeUuid     KeyType = "uuid"
	KeyTypeStandard KeyType = "standard"
)

type Redirect struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	URL  string `json:"url"`
}

type Input struct {
	KeyType   KeyType    `json:"keyType"`
	Redirects []Redirect `json:"redirects"`
}

type Item struct {
	Key       string     `json:"key"`
	Redirects []Redirect `json:"redirects"`
	Visits    int        `json:"visits"`
}
