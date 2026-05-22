package server

import "random-match/backend/internal/model"

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
	Mode   model.MatchMode `json:"mode" example:"video" enums:"video,voice"`
	Region string          `json:"region" example:"global"`
}

type waitingMatchResponse struct {
	Status string `json:"status" example:"waiting"`
}

type matchedResponse struct {
	Status    string `json:"status" example:"matched"`
	RoomID    string `json:"roomId" example:"f9b9fdc7110a4a5cb6924f2d936cd58a"`
	PeerID    string `json:"peerId" example:"5f6d8c2a7b1e4a9f8c0d2e3f5a6b7c8d"`
	Initiator bool   `json:"initiator" example:"true"`
}

type leaveMatchResponse struct {
	Status string `json:"status" example:"left"`
}

type statsResponse struct {
	Online   int `json:"online" example:"12"`
	Waiting  int `json:"waiting" example:"3"`
	Chatting int `json:"chatting" example:"8"`
}
