package service

import (
	"context"
	"time"

	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/repo"
)

// UserInfo mirrors the JSON shape that flutter_openim_sdk's UserFullInfo.fromJson expects.
type UserInfo struct {
	UserID         string `json:"userID"`
	Account        string `json:"account"`
	PhoneNumber    string `json:"phoneNumber"`
	AreaCode       string `json:"areaCode"`
	Email          string `json:"email"`
	Nickname       string `json:"nickname"`
	FaceURL        string `json:"faceURL"`
	Gender         int8   `json:"gender"`
	Birth          int64  `json:"birth"`
	Level          int8   `json:"level"`
	AllowAddFriend int8   `json:"allowAddFriend"`
	AllowBeep      int8   `json:"allowBeep"`
	AllowVibration int8   `json:"allowVibration"`
}

type UserService struct {
	users *repo.UserRepo
}

func NewUserService(users *repo.UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) FindByIDs(_ context.Context, userIDs []string) ([]UserInfo, error) {
	users, err := s.users.FindByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	return toUserInfoSlice(users), nil
}

func (s *UserService) Search(_ context.Context, keyword string, page, size int) ([]UserInfo, error) {
	offset := (page - 1) * size
	users, err := s.users.Search(keyword, offset, size)
	if err != nil {
		return nil, err
	}
	return toUserInfoSlice(users), nil
}

// Update applies only whitelisted fields to prevent mass-assignment vulnerabilities.
func (s *UserService) Update(_ context.Context, userID string, incoming map[string]any) error {
	allowed := map[string]bool{
		"nickname": true, "face_url": true, "gender": true, "birth": true,
		"email": true, "phone_number": true, "area_code": true,
		"allow_add_friend": true, "allow_beep": true, "allow_vibration": true,
	}
	safe := map[string]any{"updated_at": time.Now()}
	for k, v := range incoming {
		if allowed[k] {
			safe[k] = v
		}
	}
	return s.users.Update(userID, safe)
}

func toUserInfoSlice(users []model.User) []UserInfo {
	out := make([]UserInfo, len(users))
	for i, u := range users {
		out[i] = UserInfo{
			UserID: u.UserID, Account: derefStr(u.Account), PhoneNumber: derefStr(u.PhoneNumber),
			AreaCode: u.AreaCode, Email: derefStr(u.Email), Nickname: u.Nickname,
			FaceURL: u.FaceURL, Gender: u.Gender, Birth: u.Birth,
			Level: u.Level, AllowAddFriend: u.AllowAddFriend,
			AllowBeep: u.AllowBeep, AllowVibration: u.AllowVibration,
		}
	}
	return out
}
