package config

import (
	"fmt"
	"strings"
)

// ProfileInfo 包含 profile 的详细信息
type ProfileInfo struct {
	Name      string
	Type      string // "cloudflare" or "aws"
	IsDefault bool
	HasCreds  bool
}

// GetAllProfiles 获取所有 profile 信息
func GetAllProfiles() []ProfileInfo {
	cfg := Get()
	if cfg == nil {
		return nil
	}

	var profiles []ProfileInfo

	// Cloudflare profiles
	for name := range cfg.Cloudflare {
		profile := cfg.Cloudflare[name]
		profiles = append(profiles, ProfileInfo{
			Name:      name,
			Type:      "cloudflare",
			IsDefault: name == cfg.DefaultProfile.Cloudflare,
			HasCreds:  profile.APIToken != "",
		})
	}

	// AWS profiles
	for name := range cfg.AWS {
		profile := cfg.AWS[name]
		profiles = append(profiles, ProfileInfo{
			Name:      name,
			Type:      "aws",
			IsDefault: name == cfg.DefaultProfile.AWS,
			HasCreds:  profile.AccessKeyID != "" && profile.SecretAccessKey != "",
		})
	}

	return profiles
}

// GetProfileInfo 获取指定 profile 的信息
func GetProfileInfo(profileType, profileName string) (*ProfileInfo, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	profileType = strings.ToLower(profileType)

	switch profileType {
	case "cloudflare", "cf":
		if profileName == "" {
			profileName = cfg.DefaultProfile.Cloudflare
		}
		profile, ok := cfg.Cloudflare[profileName]
		if !ok {
			return nil, fmt.Errorf("Cloudflare profile '%s' 不存在", profileName)
		}
		return &ProfileInfo{
			Name:      profileName,
			Type:      "cloudflare",
			IsDefault: profileName == cfg.DefaultProfile.Cloudflare,
			HasCreds:  profile.APIToken != "",
		}, nil

	case "aws":
		if profileName == "" {
			profileName = cfg.DefaultProfile.AWS
		}
		profile, ok := cfg.AWS[profileName]
		if !ok {
			return nil, fmt.Errorf("AWS profile '%s' 不存在", profileName)
		}
		return &ProfileInfo{
			Name:      profileName,
			Type:      "aws",
			IsDefault: profileName == cfg.DefaultProfile.AWS,
			HasCreds:  profile.AccessKeyID != "" && profile.SecretAccessKey != "",
		}, nil

	default:
		return nil, fmt.Errorf("无效的 profile 类型: %s", profileType)
	}
}

// MaskSensitiveData 隐藏敏感信息
func MaskSensitiveData(data string) string {
	if data == "" {
		return ""
	}
	if len(data) <= 8 {
		return "****"
	}
	return data[:4] + "****" + data[len(data)-4:]
}

// GetCloudflareProfileSafe 获取 Cloudflare profile（隐藏敏感信息）
func GetCloudflareProfileSafe(profileName string) (map[string]string, error) {
	profile, err := GetCloudflareProfile(profileName)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"api_token": MaskSensitiveData(profile.APIToken),
	}, nil
}

// GetAWSProfileSafe 获取 AWS profile（隐藏敏感信息）
func GetAWSProfileSafe(profileName string) (map[string]string, error) {
	profile, err := GetAWSProfile(profileName)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_key_id":     MaskSensitiveData(profile.AccessKeyID),
		"secret_access_key": MaskSensitiveData(profile.SecretAccessKey),
		"region":            profile.Region,
	}, nil
}
