package domain

type Printer struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Transport    string         `json:"transport"`
	Status       string         `json:"status"`
	Default      bool           `json:"default,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	Connection   map[string]any `json:"connection,omitempty"`
}
