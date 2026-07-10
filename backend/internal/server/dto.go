package server

import (
	"time"

	"random-match/backend/internal/model"
)

type errorResponse struct {
	Error string `json:"error" example:"unauthorized"`
}

type healthResponse struct {
	Status string `json:"status" example:"ok"`
}

type authSessionResponse struct {
	UserID string `json:"userId" example:"4f6d8c2a7b1e4a9f8c0d2e3f5a6b7c8d"`
}

type anonymousAuthResponse struct {
	Token string     `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  model.User `json:"user"`
}

type joinMatchRequest struct {
	Mode   model.MatchMode `json:"mode" example:"video" enums:"video"`
	Region string          `json:"region" example:"global"`
	Gender string          `json:"gender" example:"everyone"`
}

type waitingMatchResponse struct {
	Status string `json:"status" example:"waiting"`
}

type matchedResponse struct {
	Status      string            `json:"status" example:"matched"`
	RoomID      string            `json:"roomId" example:"f9b9fdc7110a4a5cb6924f2d936cd58a"`
	PeerID      string            `json:"peerId" example:"5f6d8c2a7b1e4a9f8c0d2e3f5a6b7c8d"`
	PeerProfile model.UserProfile `json:"peerProfile"`
	Initiator   bool              `json:"initiator" example:"true"`
}

type leaveMatchResponse struct {
	Status string `json:"status" example:"left"`
}

type statsResponse struct {
	Online   int `json:"online" example:"12"`
	Waiting  int `json:"waiting" example:"3"`
	Chatting int `json:"chatting" example:"8"`
}

type snapshotRequest struct {
	RoomID string `json:"roomId" example:"f9b9fdc7110a4a5cb6924f2d936cd58a"`
	PeerID string `json:"peerId" example:"5f6d8c2a7b1e4a9f8c0d2e3f5a6b7c8d"`
	Mode   string `json:"mode" example:"video"`
	Image  string `json:"image" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ..."`
	Width  int    `json:"width" example:"640"`
	Height int    `json:"height" example:"480"`
}

type snapshotResponse struct {
	Status string `json:"status" example:"saved"`
	Path   string `json:"path" example:"/app/snapshots/2026-05-22/room-user.jpg"`
}

type updateProfileRequest struct {
	DisplayName  string   `json:"displayName" example:"Star Voyager"`
	Bio          string   `json:"bio" example:"喜欢深夜聊天、电影和旅行"`
	Interests    []string `json:"interests" example:"movie,music,travel"`
	Region       string   `json:"region" example:"global"`
	Gender       string   `json:"gender" example:"private"`
	AgeConfirmed bool     `json:"ageConfirmed" example:"true"`
}

type profileResponse struct {
	User model.UserProfile `json:"user"`
}

type userActionRequest struct {
	Reason string `json:"reason" example:"inappropriate behavior"`
}

type userActionResponse struct {
	Status string `json:"status" example:"ok"`
}

type blockedUserItem struct {
	User      model.UserProfile `json:"user"`
	CreatedAt time.Time         `json:"createdAt"`
}

type blockedUsersResponse struct {
	Users []blockedUserItem `json:"users"`
}

type discoverProfilesResponse struct {
	Users []model.UserProfile `json:"users"`
}

type pushSubscriptionRequest struct {
	Endpoint string               `json:"endpoint" example:"https://fcm.googleapis.com/fcm/send/..."`
	Keys     pushSubscriptionKeys `json:"keys"`
}

type pushSubscriptionKeys struct {
	Auth   string `json:"auth" example:"B8r..."`
	P256dh string `json:"p256dh" example:"BDp..."`
}

type pushSubscriptionResponse struct {
	Status string `json:"status" example:"saved"`
}

type pushTestResponse struct {
	Status string `json:"status" example:"sent"`
}

type pushDeviceTokenRequest struct {
	Token    string `json:"token" example:"fcm-registration-token"`
	Platform string `json:"platform" example:"ios" enums:"ios,android"`
}

type pushDeviceTokenResponse struct {
	Status string `json:"status" example:"saved"`
}

type commerceStatusResponse struct {
	IsMember            bool       `json:"isMember" example:"true"`
	MembershipPlan      string     `json:"membershipPlan" example:"premium_monthly"`
	MembershipExpiresAt *time.Time `json:"membershipExpiresAt,omitempty"`
	DailyLimit          int        `json:"dailyLimit" example:"10"`
	DailyUsed           int        `json:"dailyUsed" example:"3"`
	DailyRemaining      int        `json:"dailyRemaining" example:"7"`
	PriorityQueue       bool       `json:"priorityQueue" example:"true"`
	GemsBalance         int        `json:"gemsBalance" example:"300"`
}

type createPaymentRequest struct {
	Plan string `json:"plan" example:"premium_monthly"`
}

type paymentOrderResponse struct {
	Order model.PaymentOrder `json:"order"`
}
