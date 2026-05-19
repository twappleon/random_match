package model

import "time"

type User struct {
	ID          string    `bson:"_id" json:"id"`
	DisplayName string    `bson:"displayName" json:"displayName"`
	AvatarURL   string    `bson:"avatarUrl" json:"avatarUrl"`
	FirebaseUID string    `bson:"firebaseUid,omitempty" json:"firebaseUid,omitempty"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
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
