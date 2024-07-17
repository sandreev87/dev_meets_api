package service

import (
	"auth/internal/config"
	"auth/internal/domain/models"
	"auth/pkg/jwt"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
	"time"
)

type mockRepo struct{ mock.Mock }

func newMockRepo() *mockRepo { return &mockRepo{} }

func (r *mockRepo) User(ctx context.Context, uid int) (models.User, error) {
	args := r.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func (r *mockRepo) CreateUser(context.Context, models.User) (int, error) {
	panic("implement me")
}
func (r *mockRepo) UserByEmail(context.Context, string) (models.User, error) {
	panic("implement me")
}

func TestCurrentUser(t *testing.T) {
	user := &models.User{ID: 123, Email: "test@test.test"}
	conf := &config.Config{Secret: "secret_string"}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	token, _ := jwt.NewToken(*user, conf.Secret, 30*time.Minute)
	repository := newMockRepo()

	mockCall := repository.On("User", context.TODO(), user.ID).Return(*user, nil)
	result, err := NewUserService(repository, conf, logger).CurrentUser(context.TODO(), token)
	assert.Equal(t, result, user)
	require.NoError(t, err)

	result, err = NewUserService(repository, conf, logger).CurrentUser(context.TODO(), "invalid token")
	assert.Nil(t, result)
	assert.EqualError(t, err, "service.UserService.CurrentUser: token is malformed: token contains an invalid number of segments")

	mockCall.Unset()
	repository.On("User", context.TODO(), user.ID).Return(models.User{}, errors.New("some error"))
	result, err = NewUserService(repository, conf, logger).CurrentUser(context.TODO(), token)
	assert.Nil(t, result)
	assert.EqualError(t, err, "service.UserService.CurrentUser: some error")
}
