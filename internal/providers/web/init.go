package web

import (
	"flag"
	"log/slog"
	"os"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/crypto"
	"hermes-ai/internal/infras/utils"
)

func Init() {
	flag.Parse()

	if os.Getenv("SESSION_SECRET") != "" {
		if os.Getenv("SESSION_SECRET") == "random_string" {
			slog.Error("SESSION_SECRET is set to an example value, please change it to a random string.")
		} else {
			config.SessionSecret = os.Getenv("SESSION_SECRET")
		}
	}
}

func CreateRootAccountIfNeed() error {
	var user entity.User
	// if user.Status != util.UserStatusEnabled {
	if err := config.DB.First(&user).Error; err != nil {
		slog.Info("no user exists, creating a root user for you: username is root, password is 123456")
		hashedPassword, err := crypto.Password2Hash("123456")
		if err != nil {
			return err
		}
		accessToken := utils.UUID()
		if config.InitialRootAccessToken != "" {
			accessToken = config.InitialRootAccessToken
		}
		rootUser := entity.User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        entity.RoleRootUser,
			Status:      entity.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: accessToken,
			Quota:       500000000000000,
		}
		config.DB.Create(&rootUser)
		if config.InitialRootToken != "" {
			slog.Info("creating initial root token as requested")
			token := entity.Token{
				Id:             1,
				UserId:         rootUser.Id,
				Key:            config.InitialRootToken,
				Status:         entity.TokenStatusEnabled,
				Name:           "Initial Root Token",
				CreatedTime:    utils.GetTimestamp(),
				AccessedTime:   utils.GetTimestamp(),
				ExpiredTime:    -1,
				RemainQuota:    500000000000000,
				UnlimitedQuota: true,
			}
			config.DB.Create(&token)
		}
	}

	return nil
}
