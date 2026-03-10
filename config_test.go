package orbit

import (
	"testing"

	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigDefaultProxySettings(t *testing.T) {
	config := NewConfig()
	assert.Equal(t, []string{"0.0.0.0/0", "::/0"}, config.TrustedProxies)
	assert.Equal(t, []string{"X-Forwarded-For", "X-Real-IP"}, config.RemoteIPHeaders)
	assert.NotNil(t, config.CORSPolicy)
	assert.True(t, config.CORSPolicy.Enabled)
	assert.False(t, config.CORSPolicy.AllowAllOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}

func TestConfigWithProxySettingsCloneInput(t *testing.T) {
	trusted := []string{"10.0.0.0/8"}
	headers := []string{"X-Real-IP"}

	config := NewConfig().
		WithTrustedProxies(trusted).
		WithRemoteIPHeaders(headers)

	trusted[0] = "192.168.0.0/16"
	headers[0] = "X-Forwarded-For"

	assert.Equal(t, []string{"10.0.0.0/8"}, config.TrustedProxies)
	assert.Equal(t, []string{"X-Real-IP"}, config.RemoteIPHeaders)
}

func TestConfigValidationRespectsEmptyTrustedProxies(t *testing.T) {
	config := isConfigValid(&Config{
		Address:         "127.0.0.1",
		Port:            8080,
		TrustedProxies:  []string{},
		RemoteIPHeaders: []string{},
	})

	assert.Empty(t, config.TrustedProxies)
	assert.Empty(t, config.RemoteIPHeaders)
}

func TestConfigValidationKeepsDefaultTrustedProxies(t *testing.T) {
	config := isConfigValid(NewConfig())
	assert.Equal(t, []string{"0.0.0.0/0", "::/0"}, config.TrustedProxies)
}

func TestConfigWithTrustedProxiesOverridesDefaults(t *testing.T) {
	config := isConfigValid(NewConfig().WithTrustedProxies([]string{"10.0.0.0/8"}))
	assert.Equal(t, []string{"10.0.0.0/8"}, config.TrustedProxies)
}

func TestConfigWithCORSPolicyCloneInput(t *testing.T) {
	policy := com.CORSPolicy{
		Enabled:         true,
		AllowedOrigins:  []string{"https://a.example.com"},
		AllowedMethods:  []string{"GET"},
		AllowedHeaders:  []string{"Content-Type"},
		ExposeHeaders:   []string{"X-Request-Id"},
		MaxAgeSeconds:   60,
		AllowAllOrigins: false,
	}

	config := NewConfig().WithCORSPolicy(policy)
	policy.AllowedOrigins[0] = "https://changed.example.com"

	assert.NotNil(t, config.CORSPolicy)
	assert.Equal(t, []string{"https://a.example.com"}, config.CORSPolicy.AllowedOrigins)
}

func TestConfigWithCORSPolicyDisable(t *testing.T) {
	config := isConfigValid(NewConfig().WithCORSPolicy(com.CORSPolicy{Enabled: false}))
	assert.NotNil(t, config.CORSPolicy)
	assert.False(t, config.CORSPolicy.Enabled)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}

func TestConfigValidationUsesDefaultCORSPolicyWhenNil(t *testing.T) {
	config := isConfigValid(&Config{
		Address:         "127.0.0.1",
		Port:            8080,
		TrustedProxies:  []string{},
		RemoteIPHeaders: []string{},
		CORSPolicy:      nil,
	})

	assert.NotNil(t, config.CORSPolicy)
	assert.True(t, config.CORSPolicy.Enabled)
	assert.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, config.CORSPolicy.AllowedMethods)
}
