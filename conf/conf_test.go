package conf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
)

type testBusinessConfig struct {
	Room *struct {
		SettleStateTime time.Duration `yaml:"settle_state_time"`
	} `yaml:"room"`
}

type multiFormatConfig struct {
	Name   string `yaml:"name" json:"name" toml:"name"`
	Count  int    `yaml:"count" json:"count" toml:"count"`
	Nested *struct {
		Enabled bool `yaml:"enabled" json:"enabled" toml:"enabled"`
	} `yaml:"nested" json:"nested" toml:"nested"`
}

type fakeConfigClient struct {
	content     string
	err         error
	listenErr   error
	param       vo.ConfigParam
	listenParam vo.ConfigParam
	closed      bool
}

func (f *fakeConfigClient) GetConfig(param vo.ConfigParam) (string, error) {
	f.param = param
	return f.content, f.err
}

func (f *fakeConfigClient) PublishConfig(vo.ConfigParam) (bool, error) {
	return false, nil
}

func (f *fakeConfigClient) DeleteConfig(vo.ConfigParam) (bool, error) {
	return false, nil
}

func (f *fakeConfigClient) ListenConfig(param vo.ConfigParam) error {
	f.listenParam = param
	return f.listenErr
}

func (f *fakeConfigClient) CancelListenConfig(vo.ConfigParam) error {
	return nil
}

func (f *fakeConfigClient) SearchConfig(vo.SearchConfigParam) (*model.ConfigPage, error) {
	return nil, nil
}

func (f *fakeConfigClient) CloseClient() {
	f.closed = true
}

func TestLoadRequiresSystemFilename(t *testing.T) {
	resetConfigForTest(t)
	_, err := Load("")
	assert.Error(t, err)
}

func TestLoadRequiresBusinessFilenameAndTargetTogether(t *testing.T) {
	resetConfigForTest(t)

	var business testBusinessConfig
	_, err := Load("Config.yaml", WithBusiness("", &business, false))
	assert.Error(t, err, "expected error when business filename is empty")

	_, err = Load("Config.yaml", WithBusiness("business.yaml", nil, false))
	assert.Error(t, err, "expected error when business target is nil")

	_, err = Load("Config.yaml", WithBusiness("business.yaml", business, false))
	assert.Error(t, err, "expected error when business target is not a pointer")
}

func TestLoadSystemConfigOnly(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
  name: sword-ball
  env: dev
  ws_path: /game
  port: 8088
  pprof_port: 6060
`)

	_, err := Load(sysFile)
	assert.NoError(t, err)
	assert.Equal(t, 105, Application().ID)
	assert.Equal(t, "sword-ball", Application().Name)
}

func TestLoadDingTalkNotifyServiceStartupConfig(t *testing.T) {
	t.Setenv("TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_ENABLED", "true")
	t.Setenv("TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_WEBHOOK", "https://oapi.dingtalk.com/robot/send?access_token=test")
	t.Setenv("TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_SECRET", "SECtest")

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "yaml",
			filename: "Config.yaml",
			content: `
dingtalk:
  notify_service_startup:
    enabled: ${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_ENABLED}
    webhook: "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_WEBHOOK}"
    secret: "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_SECRET}"
`,
		},
		{
			name:     "json",
			filename: "Config.json",
			content: `{
  "dingtalk": {
    "notify_service_startup": {
      "enabled": ${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_ENABLED},
      "webhook": "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_WEBHOOK}",
      "secret": "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_SECRET}"
    }
  }
}`,
		},
		{
			name:     "toml",
			filename: "Config.toml",
			content: `
[dingtalk.notify_service_startup]
enabled = ${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_ENABLED}
webhook = "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_WEBHOOK}"
secret = "${TEST_DINGTALK_NOTIFY_SERVICE_STARTUP_SECRET}"
`,
		},
	}

	want := DingTalkWebhookConfig{
		Enabled: true,
		Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
		Secret:  "SECtest",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigForTest(t)
			filename := writeFile(t, t.TempDir(), tt.filename, tt.content)

			_, err := Load(filename)
			assert.NoError(t, err)
			assert.Equal(t, want, DingTalk().NotifyServiceStartup)
		})
	}
}

func TestLoadDingTalkConfigDefaults(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "dingtalk omitted", content: "application:\n  id: 105\n"},
		{name: "notify service startup empty", content: "dingtalk:\n  notify_service_startup: {}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigForTest(t)
			filename := writeFile(t, t.TempDir(), "Config.yaml", tt.content)

			_, err := Load(filename)
			assert.NoError(t, err)
			assert.Equal(t, DingTalkConfig{}, DingTalk())
		})
	}
}

func TestLoadSystemConfigWithEnvSubst(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	t.Setenv("TEST_GAME_ID", "200")
	t.Setenv("TEST_PORT", "9090")
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: ${TEST_GAME_ID}
  name: ${TEST_NAME:-sword-ball-default}
  ws_port: ${TEST_PORT}
  pprof_port: ${TEST_PPROF_PORT:-6060}
`)

	_, err := Load(sysFile)
	assert.NoError(t, err)
	assert.Equal(t, 200, Application().ID)
	assert.Equal(t, "sword-ball-default", Application().Name)
	assert.Equal(t, "9090", Application().WsPort)
	assert.Equal(t, "6060", Application().PprofPort)
}

func TestLoadSystemConfigJSONWithEnvSubst(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	t.Setenv("TEST_NAME", "json-env")
	t.Setenv("TEST_PORT", "7070")
	sysFile := writeFile(t, dir, "Config.json", `{
  "application": {
    "id": 300,
    "name": "${TEST_NAME}",
    "ws_port": "${TEST_PORT}",
    "pprof_port": "${TEST_PPROF_PORT:-6060}"
  }
}`)

	_, err := Load(sysFile)
	assert.NoError(t, err)
	assert.Equal(t, 300, Application().ID)
	assert.Equal(t, "json-env", Application().Name)
	assert.Equal(t, "7070", Application().WsPort)
	assert.Equal(t, "6060", Application().PprofPort)
}

func TestLoadSystemConfigTOMLWithEnvSubst(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	t.Setenv("TEST_NAME", "toml-env")
	t.Setenv("TEST_PORT", "5050")
	sysFile := writeFile(t, dir, "Config.toml", `
[application]
id = 400
name = "${TEST_NAME}"
ws_port = "${TEST_PORT}"
pprof_port = "${TEST_PPROF_PORT:-6060}"
`)

	_, err := Load(sysFile)
	assert.NoError(t, err)
	assert.Equal(t, 400, Application().ID)
	assert.Equal(t, "toml-env", Application().Name)
	assert.Equal(t, "5050", Application().WsPort)
	assert.Equal(t, "6060", Application().PprofPort)
}

func TestLoadBusinessFromLocal(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
  name: sword-ball
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 30s
`)
	var business testBusinessConfig

	_, err := Load(sysFile, WithBusiness(businessFile, &business, false))
	assert.NoError(t, err)
	assert.NotNil(t, business.Room)
	assert.Equal(t, 30*time.Second, business.Room.SettleStateTime)
}

func TestLoadBusinessFromLocalWhenConfigCenterInvalid(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 45s
`)
	var business testBusinessConfig

	_, err := Load(sysFile, WithBusiness(businessFile, &business, true))
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Second, business.Room.SettleStateTime)
}

func TestLoadBusinessFromNacosDoesNotOverwriteSystemConfig(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
  name: sword-ball
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.yaml
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 10s
`)
	fake := &fakeConfigClient{content: `
application:
  id: 999
room:
  settle_state_time: 60s
`}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business testBusinessConfig

	closeFn, err := Load(sysFile, WithBusiness(businessFile, &business, false))
	assert.NoError(t, err)
	closeFn()
	assert.Equal(t, "business.yaml", fake.param.DataId)
	assert.Equal(t, "DEFAULT_GROUP", fake.param.Group)
	assert.True(t, fake.closed, "expected nacos client to be closed")
	assert.Equal(t, 60*time.Second, business.Room.SettleStateTime)
	assert.Equal(t, 105, Application().ID)
}

func TestLoadBusinessFromNacosError(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.yaml
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 10s
`)
	restore := replaceConfigClientFactory(t, &fakeConfigClient{err: errors.New("boom")}, nil)
	defer restore()
	var business testBusinessConfig

	_, err := Load(sysFile, WithBusiness(businessFile, &business, false))
	assert.Error(t, err)
}

func TestLoadBusinessFromNacosWatch(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.yaml
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 10s
`)
	fake := &fakeConfigClient{content: `
room:
  settle_state_time: 60s
`}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business testBusinessConfig

	closeFn, err := Load(sysFile, WithBusiness(businessFile, &business, true))
	assert.NoError(t, err)
	defer closeFn()
	assert.Equal(t, "business.yaml", fake.listenParam.DataId)
	assert.Equal(t, "DEFAULT_GROUP", fake.listenParam.Group)
	assert.NotNil(t, fake.listenParam.OnChange, "expected OnChange callback")
	assert.False(t, fake.closed, "expected watched nacos client to stay open")

	fake.listenParam.OnChange("dev", "DEFAULT_GROUP", "business.yaml", `
room:
  settle_state_time: 90s
`)
	assert.Equal(t, 90*time.Second, business.Room.SettleStateTime)

	closeFn()
	assert.True(t, fake.closed, "expected close function to close watched nacos client")
}

func TestLoadBusinessFromNacosWatchErrorClosesClient(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.yaml
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.yaml", `
room:
  settle_state_time: 10s
`)
	fake := &fakeConfigClient{
		content: `
room:
  settle_state_time: 60s
`,
		listenErr: errors.New("listen boom"),
	}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business testBusinessConfig

	_, err := Load(sysFile, WithBusiness(businessFile, &business, true))
	assert.Error(t, err)
	assert.True(t, fake.closed, "expected client to be closed when listen fails")
}

func TestUnmarshalConfigYAML(t *testing.T) {
	var cfg multiFormatConfig
	content := `name: sword-ball
count: 10
nested:
  enabled: true
`
	err := unmarshalConfig("business.yaml", []byte(content), &cfg)
	assert.NoError(t, err)
	assertMultiFormat(t, cfg, "sword-ball", 10, true)
}

func TestUnmarshalConfigYML(t *testing.T) {
	var cfg multiFormatConfig
	content := `name: test-yml
count: 5
`
	err := unmarshalConfig("business.yml", []byte(content), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "test-yml", cfg.Name)
	assert.Equal(t, 5, cfg.Count)
}

func TestUnmarshalConfigJSON(t *testing.T) {
	var cfg multiFormatConfig
	content := `{"name": "sword-ball", "count": 20, "nested": {"enabled": true}}`
	err := unmarshalConfig("business.json", []byte(content), &cfg)
	assert.NoError(t, err)
	assertMultiFormat(t, cfg, "sword-ball", 20, true)
}

func TestUnmarshalConfigTOML(t *testing.T) {
	var cfg multiFormatConfig
	content := `name = "sword-ball"
count = 30

[nested]
enabled = true
`
	err := unmarshalConfig("business.toml", []byte(content), &cfg)
	assert.NoError(t, err)
	assertMultiFormat(t, cfg, "sword-ball", 30, true)
}

func TestUnmarshalConfigDefaultsToYAML(t *testing.T) {
	var cfg multiFormatConfig
	content := `name: default-yaml
count: 1
`
	err := unmarshalConfig("business.conf", []byte(content), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "default-yaml", cfg.Name)
	assert.Equal(t, 1, cfg.Count)
}

func TestUnmarshalConfigCaseInsensitive(t *testing.T) {
	var cfg multiFormatConfig
	content := `{"name": "upper-json", "count": 7}`
	err := unmarshalConfig("business.JSON", []byte(content), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "upper-json", cfg.Name)
}

func TestLoadBusinessFromNacosJSON(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.json
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.json", `{"name": "local"}`)
	fake := &fakeConfigClient{content: `{"name": "nacos-json", "count": 42}`}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business multiFormatConfig

	closeFn, err := Load(sysFile, WithBusiness(businessFile, &business, false))
	assert.NoError(t, err)
	closeFn()
	assert.Equal(t, "nacos-json", business.Name)
	assert.Equal(t, 42, business.Count)
}

func TestLoadBusinessFromNacosTOML(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.toml
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.toml", `name = "local"`)
	fake := &fakeConfigClient{content: `name = "nacos-toml"
count = 99`}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business multiFormatConfig

	closeFn, err := Load(sysFile, WithBusiness(businessFile, &business, false))
	assert.NoError(t, err)
	closeFn()
	assert.Equal(t, "nacos-toml", business.Name)
	assert.Equal(t, 99, business.Count)
}

func TestLoadBusinessFromNacosWatchJSON(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	sysFile := writeFile(t, dir, "Config.yaml", `
application:
  id: 105
config_center:
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business.json
  group: DEFAULT_GROUP
`)
	businessFile := writeFile(t, dir, "business.json", `{"name": "local"}`)
	fake := &fakeConfigClient{content: `{"name": "v1"}`}
	restore := replaceConfigClientFactory(t, fake, nil)
	defer restore()
	var business multiFormatConfig

	closeFn, err := Load(sysFile, WithBusiness(businessFile, &business, true))
	assert.NoError(t, err)
	defer closeFn()
	assert.Equal(t, "v1", business.Name)

	fake.listenParam.OnChange("dev", "DEFAULT_GROUP", "business.json", `{"name": "v2", "count": 100}`)
	assert.Equal(t, "v2", business.Name)
	assert.Equal(t, 100, business.Count)
}

func assertMultiFormat(t *testing.T, cfg multiFormatConfig, wantName string, wantCount int, wantEnabled bool) {
	t.Helper()
	assert.Equal(t, wantName, cfg.Name)
	assert.Equal(t, wantCount, cfg.Count)
	assert.NotNil(t, cfg.Nested)
	assert.Equal(t, wantEnabled, cfg.Nested.Enabled)
}

func resetConfigForTest(t *testing.T) {
	t.Helper()
	if cfgClient != nil {
		cfgClient.CloseClient()
		cfgClient = nil
	}
	cfg = Config{}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	filename := filepath.Join(dir, name)
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
	return filename
}

func replaceConfigClientFactory(t *testing.T, client config_client.IConfigClient, err error) func() {
	t.Helper()
	previous := buildConfigClientFunc
	buildConfigClientFunc = func(*NacosConfig) (config_client.IConfigClient, error) {
		return client, err
	}
	return func() {
		buildConfigClientFunc = previous
	}
}
