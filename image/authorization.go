package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	ggoogle "github.com/google/go-containerregistry/pkg/v1/google"
)

type basicAuthPassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GenAuthenticator(input string) (authn.Authenticator, error) {
	input = strings.TrimSpace(input)

	if len(input) == 0 {
		return authn.Anonymous, nil
	}

	// Case 2: 形如 user:password
	if strings.Contains(input, ":") && !strings.HasPrefix(input, "{") {
		return GenBasicUserPassword(input)
	}

	// Case 3: 可能是 JSON 文件路径（Docker config 或 GCP key）
	if _, err := os.Stat(input); err == nil {
		absPath, _ := filepath.Abs(input)
		return GenGSAAuthenticatorFromFile(absPath)
	}

	// Case 4: 直接传入 Google serviceAccount JSON 内容
	if strings.HasPrefix(input, "{") && strings.Contains(input, "\"type\":") {
		var keyMap map[string]interface{}
		if err := json.Unmarshal([]byte(input), &keyMap); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return ggoogle.NewJSONKeyAuthenticator(input), nil
	}

	return nil, errors.New("not support authenticator type, please use user@password format or google cloud service account json file path")
}

func GenBasicUserPassword(authMeta string) (authn.Authenticator, error) {
	authInfo := basicAuthPassword{}
	err := json.Unmarshal([]byte(authMeta), &authInfo)
	if err != nil {
		return nil, err
	}
	return &authn.Basic{Username: authInfo.Username, Password: authInfo.Password}, nil
}

func GenGSAAuthenticatorFromFile(authFilePath string) (authn.Authenticator, error) {
	authFile, err := os.Open(authFilePath)
	if err != nil {
		return nil, err
	}
	authContent, err := io.ReadAll(authFile)
	if err != nil {
		return nil, err
	}
	return ggoogle.NewJSONKeyAuthenticator(string(authContent)), nil
}

func GenGSAAuthenticatorFromJSON(authJSON string) (authn.Authenticator, error) {
	return ggoogle.NewJSONKeyAuthenticator(authJSON), nil
}
