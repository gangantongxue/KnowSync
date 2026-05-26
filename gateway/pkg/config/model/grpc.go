package model

// GRPCClientsCfg gRPC 多客户端配置
// key 为服务名（如 user_server），value 为目标地址
type GRPCClientsCfg struct {
	Targets map[string]string `yaml:"targets"`
}
