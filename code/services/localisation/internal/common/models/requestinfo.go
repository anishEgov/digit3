package models

import "time"

// RequestInfo is common structure used across all API requests for DIGIT platform
type RequestInfo struct {
	APIId       string    `json:"apiId"`
	Ver         string    `json:"ver"`
	Ts          time.Time `json:"ts"`
	Action      string    `json:"action"`
	Did         string    `json:"did"`
	Key         string    `json:"key"`
	MsgId       string    `json:"msgId"`
	AuthToken   string    `json:"authToken"`
	RequesterId string    `json:"requesterId,omitempty"`
	UserInfo    *UserInfo `json:"userInfo,omitempty"`
}

// UserInfo represents the authenticated user information
type UserInfo struct {
	ID       int64  `json:"id"`
	UUID     string `json:"uuid,omitempty"`
	UserName string `json:"userName,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
	Type     string `json:"type,omitempty"`
	TenantId string `json:"tenantId,omitempty"`
	Roles    []Role `json:"roles,omitempty"`
}

// Role represents a user's role
type Role struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	TenantId    string `json:"tenantId,omitempty"`
}
