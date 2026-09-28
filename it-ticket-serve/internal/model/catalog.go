package model

type TemplateField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
}

type TicketTemplate struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Category  string          `json:"category"`
	TitleHint string          `json:"title_hint"`
	Hint      string          `json:"hint"`
	Icon      string          `json:"icon"`
	SortOrder int             `json:"sort_order"`
	Enabled   bool            `json:"enabled"`
	Fields    []TemplateField `json:"fields"`
}

type CannedReply struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
}

type TicketTemplateList struct {
	Items []TicketTemplate `json:"items"`
}

type CannedReplyList struct {
	Items []CannedReply `json:"items"`
}

type TemplateInput struct {
	Name      string          `json:"name"`
	Category  string          `json:"category"`
	TitleHint string          `json:"title_hint"`
	Hint      string          `json:"hint"`
	Icon      string          `json:"icon"`
	SortOrder int             `json:"sort_order"`
	Enabled   *bool           `json:"enabled"`
	Fields    []TemplateField `json:"fields"`
}

type CannedReplyInput struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}
