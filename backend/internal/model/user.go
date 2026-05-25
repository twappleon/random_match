package model

import "time"

type User struct {
	ID           string    `bson:"_id" json:"id"`
	DisplayName  string    `bson:"displayName" json:"displayName"`
	AvatarURL    string    `bson:"avatarUrl" json:"avatarUrl"`
	Bio          string    `bson:"bio,omitempty" json:"bio,omitempty"`
	Interests    []string  `bson:"interests,omitempty" json:"interests,omitempty"`
	AgeConfirmed bool      `bson:"ageConfirmed" json:"ageConfirmed"`
	FirebaseUID  string    `bson:"firebaseUid,omitempty" json:"firebaseUid,omitempty"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt" json:"updatedAt"`
}

type UserProfile struct {
	ID           string   `bson:"_id" json:"id"`
	DisplayName  string   `bson:"displayName" json:"displayName"`
	AvatarURL    string   `bson:"avatarUrl" json:"avatarUrl"`
	Bio          string   `bson:"bio,omitempty" json:"bio,omitempty"`
	Interests    []string `bson:"interests,omitempty" json:"interests,omitempty"`
	AgeConfirmed bool     `bson:"ageConfirmed" json:"ageConfirmed"`
}

type UserBlock struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"userId" json:"userId"`
	BlockedID string    `bson:"blockedId" json:"blockedId"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type UserReport struct {
	ID         string    `bson:"_id" json:"id"`
	ReporterID string    `bson:"reporterId" json:"reporterId"`
	TargetID   string    `bson:"targetId" json:"targetId"`
	Reason     string    `bson:"reason" json:"reason"`
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
}

type MatchMode string

const (
	MatchModeVideo MatchMode = "video"
	MatchModeVoice MatchMode = "voice"
)

type MatchTicket struct {
	UserID    string    `json:"userId"`
	Mode      MatchMode `json:"mode"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"createdAt"`
}
