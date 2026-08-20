package model

import "time"

type User struct {
	ID                  string     `bson:"_id" json:"id"`
	DisplayName         string     `bson:"displayName" json:"displayName"`
	AvatarURL           string     `bson:"avatarUrl" json:"avatarUrl"`
	Bio                 string     `bson:"bio,omitempty" json:"bio,omitempty"`
	Interests           []string   `bson:"interests,omitempty" json:"interests,omitempty"`
	Region              string     `bson:"region,omitempty" json:"region,omitempty"`
	Latitude            *float64   `bson:"latitude,omitempty" json:"-"`
	Longitude           *float64   `bson:"longitude,omitempty" json:"-"`
	LocationUpdatedAt   *time.Time `bson:"locationUpdatedAt,omitempty" json:"locationUpdatedAt,omitempty"`
	Gender              string     `bson:"gender,omitempty" json:"gender,omitempty"`
	Language            string     `bson:"language,omitempty" json:"language,omitempty"`
	GemsBalance         int        `bson:"gemsBalance,omitempty" json:"gemsBalance,omitempty"`
	AgeConfirmed        bool       `bson:"ageConfirmed" json:"ageConfirmed"`
	MembershipPlan      string     `bson:"membershipPlan,omitempty" json:"membershipPlan,omitempty"`
	MembershipExpiresAt *time.Time `bson:"membershipExpiresAt,omitempty" json:"membershipExpiresAt,omitempty"`
	FirebaseUID         string     `bson:"firebaseUid,omitempty" json:"firebaseUid,omitempty"`
	CreatedAt           time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time  `bson:"updatedAt" json:"updatedAt"`
}

type UserProfile struct {
	ID                  string     `bson:"_id" json:"id"`
	DisplayName         string     `bson:"displayName" json:"displayName"`
	AvatarURL           string     `bson:"avatarUrl" json:"avatarUrl"`
	Bio                 string     `bson:"bio,omitempty" json:"bio,omitempty"`
	Interests           []string   `bson:"interests,omitempty" json:"interests,omitempty"`
	Region              string     `bson:"region,omitempty" json:"region,omitempty"`
	DistanceKm          *float64   `bson:"-" json:"distanceKm,omitempty"`
	LocationUpdatedAt   *time.Time `bson:"locationUpdatedAt,omitempty" json:"locationUpdatedAt,omitempty"`
	Gender              string     `bson:"gender,omitempty" json:"gender,omitempty"`
	Language            string     `bson:"language,omitempty" json:"language,omitempty"`
	TrustBadge          bool       `bson:"-" json:"trustBadge"`
	AgeConfirmed        bool       `bson:"ageConfirmed" json:"ageConfirmed"`
	MembershipPlan      string     `bson:"membershipPlan,omitempty" json:"membershipPlan,omitempty"`
	MembershipExpiresAt *time.Time `bson:"membershipExpiresAt,omitempty" json:"membershipExpiresAt,omitempty"`
	IsMember            bool       `bson:"-" json:"isMember"`
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

type UserFollow struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"userId" json:"userId"`
	FollowID  string    `bson:"followId" json:"followId"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type DirectMessage struct {
	ID         string    `bson:"_id" json:"id"`
	SenderID   string    `bson:"senderId" json:"senderId"`
	ReceiverID string    `bson:"receiverId" json:"receiverId"`
	Text       string    `bson:"text" json:"text"`
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
}

type MatchMode string

const (
	MatchModeVideo MatchMode = "video"
	MatchModeVoice MatchMode = "voice"
)

type MatchTicket struct {
	UserID           string    `json:"userId"`
	Mode             MatchMode `json:"mode"`
	Region           string    `json:"region"`
	GenderPreference string    `json:"genderPreference"`
	Language         string    `json:"language"`
	Interests        []string  `json:"interests"`
	CreatedAt        time.Time `json:"createdAt"`
}

type PaymentOrder struct {
	ID        string     `bson:"_id" json:"id"`
	UserID    string     `bson:"userId" json:"userId"`
	Plan      string     `bson:"plan" json:"plan"`
	Amount    int        `bson:"amount" json:"amount"`
	Currency  string     `bson:"currency" json:"currency"`
	Status    string     `bson:"status" json:"status"`
	CreatedAt time.Time  `bson:"createdAt" json:"createdAt"`
	PaidAt    *time.Time `bson:"paidAt,omitempty" json:"paidAt,omitempty"`
}

type AppleIAPPurchase struct {
	ID                    string    `bson:"_id" json:"id"`
	UserID                string    `bson:"userId" json:"userId"`
	ProductID             string    `bson:"productId" json:"productId"`
	TransactionID         string    `bson:"transactionId" json:"transactionId"`
	OriginalTransactionID string    `bson:"originalTransactionId,omitempty" json:"originalTransactionId,omitempty"`
	Environment           string    `bson:"environment" json:"environment"`
	ExpiresAt             time.Time `bson:"expiresAt" json:"expiresAt"`
	VerifiedAt            time.Time `bson:"verifiedAt" json:"verifiedAt"`
	Source                string    `bson:"source,omitempty" json:"source,omitempty"`
}
