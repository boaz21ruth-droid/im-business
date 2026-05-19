package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/pkg/resp"
)

type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	// Map Flutter camelCase field names to DB snake_case column names
	fields := map[string]any{}
	remap := map[string]string{
		"nickname": "nickname", "faceURL": "face_url", "gender": "gender",
		"birth": "birth", "email": "email", "phoneNumber": "phone_number",
		"areaCode": "area_code", "allowAddFriend": "allow_add_friend",
		"allowBeep": "allow_beep", "allowVibration": "allow_vibration",
	}
	for flutter, db := range remap {
		if v, ok := req[flutter]; ok {
			fields[db] = v
		}
	}
	if err := h.svc.Update(c.Request.Context(), userID, fields); err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *UserHandler) FindUsersFullInfo(c *gin.Context) {
	var req struct {
		UserIDs    []string `json:"userIDs"`
		Pagination struct {
			PageNumber int `json:"pageNumber"`
			ShowNumber int `json:"showNumber"`
		} `json:"pagination"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	users, err := h.svc.FindByIDs(c.Request.Context(), req.UserIDs)
	if err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"users": users})
}

func (h *UserHandler) SearchUsersFullInfo(c *gin.Context) {
	var req struct {
		Keyword    string `json:"keyword"`
		Pagination struct {
			PageNumber int `json:"pageNumber"`
			ShowNumber int `json:"showNumber"`
		} `json:"pagination"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	page, size := req.Pagination.PageNumber, req.Pagination.ShowNumber
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	users, err := h.svc.Search(c.Request.Context(), req.Keyword, page, size)
	if err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"users": users})
}

// SearchFriendInfo reuses user search for MVP. OpenIM SDK already filters actual friends.
func (h *UserHandler) SearchFriendInfo(c *gin.Context) {
	h.SearchUsersFullInfo(c)
}
