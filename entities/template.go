package entities

import "time"

type Template struct {
	Model
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	HTMLPart string `json:"html_part" gorm:"-"`
	TextPart string `json:"text_part"`
	Subject  string `json:"subject"`
}

type TemplateCollection struct {
	NextToken  string         `json:"next_token"`
	Collection []TemplateMeta `json:"collection"`
}

type TemplateMeta struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}
