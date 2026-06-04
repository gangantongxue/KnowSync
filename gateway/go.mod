module github.com/gangantongxue/knowsync/gateway

go 1.26.0

require (
	github.com/cloudwego/hertz v0.10.4
	github.com/gangantongxue/knowsync/ks-proto v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/hertz-contrib/cors v0.1.0
	github.com/spf13/viper v1.21.0
	google.golang.org/grpc v1.81.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require go.uber.org/atomic v1.11.0 // indirect

replace github.com/gangantongxue/knowsync/ks-proto => ../ks-proto

require (
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.2 // indirect
	github.com/bytedance/sonic/loader v0.5.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/cloudwego/gopkg v0.2.0 // indirect
	github.com/cloudwego/netpoll v0.7.2 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/redis/go-redis/v9 v9.20.0
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.27.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
