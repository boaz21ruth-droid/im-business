package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"

	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/repo"
	codepkg "github.com/web1/im-business/pkg/code"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
)

type AccountService struct {
	users  *repo.UserRepo
	openim *OpenIMClient
	jwt    *jwtpkg.Service
	code   *codepkg.Service
	node   *snowflake.Node
}

func NewAccountService(
	users *repo.UserRepo,
	openim *OpenIMClient,
	jwt *jwtpkg.Service,
	code *codepkg.Service,
) (*AccountService, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	return &AccountService{users: users, openim: openim, jwt: jwt, code: code, node: node}, nil
}

type LoginResult struct {
	UserID    string
	IMToken   string
	ChatToken string
}

type RegisterInput struct {
	Account     string
	PhoneNumber string
	AreaCode    string
	Email       string
	Password    string // MD5 from client
	Nickname    string
	FaceURL     string
	Gender      int8
	Birth       int64
	VerifyCode  string
	Platform    int
}

type LoginInput struct {
	Account  string // may be email, phone, or account handle
	Password string // MD5 from client
	Platform int
}

func (s *AccountService) SendCode(ctx context.Context, target string) error {
	if target == "" {
		return errors.New("target required")
	}
	return s.code.Send(ctx, target)
}

func (s *AccountService) VerifyCode(ctx context.Context, target, submitted string) error {
	return s.code.Verify(ctx, target, submitted)
}

func (s *AccountService) Register(ctx context.Context, in RegisterInput) (*LoginResult, error) {
	target := firstNonEmpty(in.Email, in.PhoneNumber, in.Account)
	if target == "" {
		return nil, errors.New("email, phone, or account required")
	}
	if err := s.code.Verify(ctx, target, in.VerifyCode); err != nil {
		return nil, fmt.Errorf("verification: %w", err)
	}

	exists, err := s.users.ExistsByAccount(in.Account, in.Email, in.PhoneNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("account already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := fmt.Sprintf("%d", s.node.Generate().Int64())
	nickname := in.Nickname
	if nickname == "" {
		nickname = "User_" + userID[len(userID)-6:]
	}

	if err := s.openim.RegisterUser(userID, nickname, in.FaceURL); err != nil {
		return nil, fmt.Errorf("openim register: %w", err)
	}

	u := &model.User{
		UserID:      userID,
		Account:     nullableStr(in.Account),
		PhoneNumber: nullableStr(in.PhoneNumber),
		AreaCode:    in.AreaCode,
		Email:       nullableStr(in.Email),
		Password:    string(hash),
		Nickname:    nickname,
		FaceURL:     in.FaceURL,
		Gender:      in.Gender,
		Birth:       in.Birth,
	}
	if err := s.users.Create(u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// All steps succeeded — invalidate the one-time code.
	s.code.Consume(ctx, target)

	return s.buildResult(userID, in.Platform)
}

func (s *AccountService) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if in.Account == "" {
		return nil, errors.New("account required")
	}
	u, err := s.users.FindByAccount(in.Account)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("account not found")
	}
	if u.Status != 1 {
		return nil, errors.New("account is banned")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, errors.New("wrong password")
	}
	return s.buildResult(u.UserID, in.Platform)
}

func (s *AccountService) ResetPassword(ctx context.Context, target, verifyCode, newPwd string) error {
	if err := s.code.Verify(ctx, target, verifyCode); err != nil {
		return fmt.Errorf("verification: %w", err)
	}
	u, err := s.users.FindByAccount(target)
	if err != nil || u == nil {
		return errors.New("user not found")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err := s.users.Update(u.UserID, map[string]any{"password": string(hash), "updated_at": time.Now()}); err != nil {
		return err
	}
	s.code.Consume(ctx, target)
	return nil
}

func (s *AccountService) ChangePassword(ctx context.Context, userID, currentPwd, newPwd string) error {
	u, err := s.users.FindByUserID(userID)
	if err != nil || u == nil {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPwd)); err != nil {
		return errors.New("wrong current password")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	return s.users.Update(userID, map[string]any{"password": string(hash), "updated_at": time.Now()})
}

func (s *AccountService) buildResult(userID string, platform int) (*LoginResult, error) {
	imToken, err := s.openim.GetUserToken(userID, platform)
	if err != nil {
		return nil, fmt.Errorf("get im token: %w", err)
	}
	chatToken, err := s.jwt.Sign(userID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{UserID: userID, IMToken: imToken, ChatToken: chatToken}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
