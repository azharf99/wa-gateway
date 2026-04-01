package domain

import "context"

// Request payload untuk pengiriman pesan
type SendMessageReq struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Response standar
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type GroupInfo struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

// WhatsAppRepository mendefinisikan kontrak untuk berinteraksi dengan engine WA
type WhatsAppRepository interface {
	Connect() error
	IsConnected() bool
	SendTextMessage(ctx context.Context, jid string, text string) (string, error)
	SendMediaMessage(ctx context.Context, jid string, req SendMediaReq) (string, error)
	GetJoinedGroups(ctx context.Context) ([]GroupInfo, error)
}

// WhatsAppUsecase mendefinisikan kontrak untuk logic bisnis
type WhatsAppUsecase interface {
	CheckStatus() string
	SendMessage(ctx context.Context, req SendMessageReq) (string, error)
	BroadcastMessages(req BroadcastReq)
	SendMedia(ctx context.Context, req SendMediaReq) (string, error)
	GetJoinedGroups(ctx context.Context) ([]GroupInfo, error)
}

// Struct untuk satu penerima dalam antrean broadcast
type BroadcastRecipient struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Struct untuk payload utama
type BroadcastReq struct {
	Recipients []BroadcastRecipient `json:"recipients" binding:"required"`
}

type SendMediaReq struct {
	To        string
	IsGroup   bool
	FileBytes []byte
	FileName  string
	MimeType  string
	Caption   string
	MediaType string // "document", "image", "video"
}
