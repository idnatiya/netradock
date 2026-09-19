package model

type WebResponse[T any] struct {
	Data T `json:"data"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,max=255"`
	Password string `json:"password" validate:"required,max=255"`
}

type UserResponse struct {
	Username string `json:"username"`
}

type PullImageRequest struct {
	Image string `json:"image" validate:"required,max=255"`
}

type PortResponse struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
}

type ContainerResponse struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Image   string         `json:"image"`
	State   string         `json:"state"`
	Status  string         `json:"status"`
	Created int64          `json:"created"`
	Ports   []PortResponse `json:"ports"`
	Project string         `json:"project,omitempty"`
}

type ImageResponse struct {
	ID         string   `json:"id"`
	Tags       []string `json:"tags"`
	Size       int64    `json:"size"`
	Created    int64    `json:"created"`
	Containers int64    `json:"containers"`
}

type VolumeResponse struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
	CreatedAt  string `json:"created_at"`
}

type NetworkResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Driver  string `json:"driver"`
	Scope   string `json:"scope"`
	Created int64  `json:"created"`
}

type SystemResponse struct {
	Name              string `json:"name"`
	ServerVersion     string `json:"server_version"`
	OperatingSystem   string `json:"operating_system"`
	KernelVersion     string `json:"kernel_version"`
	Architecture      string `json:"architecture"`
	NCPU              int    `json:"ncpu"`
	MemTotal          int64  `json:"mem_total"`
	Containers        int    `json:"containers"`
	ContainersRunning int    `json:"containers_running"`
	ContainersPaused  int    `json:"containers_paused"`
	ContainersStopped int    `json:"containers_stopped"`
	Images            int    `json:"images"`
}

type StatsResponse struct {
	CPUPercent float64 `json:"cpu_percent"`
	MemUsage   uint64  `json:"mem_usage"`
	MemLimit   uint64  `json:"mem_limit"`
	NetRx      uint64  `json:"net_rx"`
	NetTx      uint64  `json:"net_tx"`
}
