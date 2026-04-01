package whatsapp

import (
	"context"
	"fmt"
	"os"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite"
)

type whatsmeowRepo struct {
	client *whatsmeow.Client
}

func NewWhatsmeowRepository() domain.WhatsAppRepository {
	dbLog := waLog.Stdout("Database", "WARN", true)

	ginMode := os.Getenv("GIN_MODE")
	var dsn string
	if ginMode == "release" {
		dsn = "file:data/sessions.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	} else {
		dsn = "file:sessions.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	}

	container, err := sqlstore.New(context.Background(), "sqlite", dsn, dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &whatsmeowRepo{
		client: client,
	}
}

func (r *whatsmeowRepo) Connect() error {
	if r.client.Store.ID == nil {
		qrChan, _ := r.client.GetQRChannel(context.Background())
		err := r.client.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login WhatsApp Sukses!")
			}
		}
	} else {
		return r.client.Connect()
	}
	return nil
}

func (r *whatsmeowRepo) IsConnected() bool {
	return r.client.IsConnected() && r.client.IsLoggedIn()
}

func (r *whatsmeowRepo) SendTextMessage(ctx context.Context, jidStr string, text string) (string, error) {
	jid, _ := types.ParseJID(jidStr)

	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	resp, err := r.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (r *whatsmeowRepo) SendMediaMessage(ctx context.Context, jidStr string, req domain.SendMediaReq) (string, error) {
	jid, _ := types.ParseJID(jidStr)

	// Tentukan tipe media whatsmeow
	var waMediaType whatsmeow.MediaType
	switch req.MediaType {
	case "image":
		waMediaType = whatsmeow.MediaImage
	case "video":
		waMediaType = whatsmeow.MediaVideo
	default:
		waMediaType = whatsmeow.MediaDocument
	}

	// 1. Upload file ke server WhatsApp
	uploaded, err := r.client.Upload(ctx, req.FileBytes, waMediaType)
	if err != nil {
		return "", fmt.Errorf("gagal upload media ke WA: %v", err)
	}

	// 2. Rakit pesan protobuf dengan meta-data hasil upload
	msg := &waE2E.Message{}

	switch req.MediaType {
	case "document":
		msg.DocumentMessage = &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(req.FileBytes))),
			Title:         proto.String(req.FileName),
			Mimetype:      proto.String(req.MimeType),
			Caption:       proto.String(req.Caption),
		}
	case "image":
		msg.ImageMessage = &waE2E.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(req.FileBytes))),
			Mimetype:      proto.String(req.MimeType),
			Caption:       proto.String(req.Caption),
		}
	case "video":
		msg.VideoMessage = &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(req.FileBytes))),
			Mimetype:      proto.String(req.MimeType),
			Caption:       proto.String(req.Caption),
		}
	}

	// 3. Eksekusi pengiriman
	resp, err := r.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (r *whatsmeowRepo) GetJoinedGroups(ctx context.Context) ([]domain.GroupInfo, error) {
	groups, err := r.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, err
	}

	var result []domain.GroupInfo
	for _, group := range groups {
		result = append(result, domain.GroupInfo{
			JID:  group.JID.User,
			Name: group.Name,
		})
	}
	return result, nil
}
