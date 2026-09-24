package model

import (
	"time"
)

// ProjectStats represents the RPG stats for a project
type ProjectStats struct {
	HP      int `json:"hp" bson:"hp"`
	Attack  int `json:"attack" bson:"attack"`
	Defense int `json:"defense" bson:"defense"`
	Speed   int `json:"speed" bson:"speed"`
}

// Project represents a portfolio project item
type Project struct {
	ID          string       `json:"id" bson:"_id,omitempty"`
	Slug        string       `json:"slug" bson:"slug"`
	Title       string       `json:"title" bson:"title"`
	Summary     string       `json:"summary" bson:"summary"`
	Description string       `json:"description" bson:"description"`
	Category    string       `json:"category" bson:"category"`
	BoxCategory string       `json:"boxCategory" bson:"boxCategory"`
	TechStack   []string     `json:"techStack" bson:"techStack"`
	DemoURL     string       `json:"demoUrl" bson:"demoUrl"`
	RepoURL     string       `json:"repoUrl" bson:"repoUrl"`
	CoverImage  string       `json:"coverImage" bson:"coverImage"`
	BallType    string       `json:"ballType" bson:"ballType"`
	TypeBadge   string       `json:"typeBadge" bson:"typeBadge"`
	Level       int          `json:"level" bson:"level"`
	Featured    bool         `json:"featured" bson:"featured"`
	Stats       ProjectStats `json:"stats" bson:"stats"`
	FlavorText  string       `json:"flavorText" bson:"flavorText"`
	KeyFeatures []string     `json:"keyFeatures" bson:"keyFeatures"`
	CreatedAt   time.Time    `json:"createdAt" bson:"createdAt"`
}

