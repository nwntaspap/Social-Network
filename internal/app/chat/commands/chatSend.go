package chatcommands

import (
	"context"

	"social-network/internal/domain/chat"

	chatapp "social-network/internal/app/chat"
)

type SendChatRequest struct {
	ChatID          string
	SenderID        string
	Content         string
	ClientMessageID string
	RequestID       string
}

type SendChatResponse struct {
	Msg         *chat.Message
	RecipientID string
}
type SendChatHandler interface {
	Handle(ctx context.Context, req SendChatRequest) (res SendChatResponse, err error)
}

type sendChatHandler struct {
	chatRepo    chat.Repository
	broadcaster chatapp.Broadcaster
}

func NewSendChatHandler(chatRepo chat.Repository, broadcaster chatapp.Broadcaster) SendChatHandler {
	return &sendChatHandler{
		chatRepo:    chatRepo,
		broadcaster: broadcaster,
	}
}

func (h *sendChatHandler) Handle(ctx context.Context, req SendChatRequest) (res SendChatResponse, err error) {
	msg, err := h.chatRepo.SendMessage(ctx, req.ChatID, req.SenderID, req.Content, req.ClientMessageID)
	if err != nil {
		return SendChatResponse{}, err
	}

	c, err := h.chatRepo.GetChat(ctx, req.ChatID)
	if err != nil {
		return SendChatResponse{}, err
	}
	if c.UserHighID == req.SenderID {
		res.RecipientID = c.UserLowID
	} else {
		res.RecipientID = c.UserHighID
	}
	res.Msg = msg

	h.broadcaster.SendToUser(res.RecipientID, req.RequestID, msg)

	return res, nil
}
