package converter

import (
	"encoding/json"
	"strings"

	"github.com/idnatiya/netradock/internal/model"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/api/types/volume"
)

func ContainerToResponse(c container.Summary) model.ContainerResponse {
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}
	ports := make([]model.PortResponse, 0, len(c.Ports))
	for _, p := range c.Ports {
		ip := ""
		if p.IP.IsValid() {
			ip = p.IP.String()
		}
		ports = append(ports, model.PortResponse{IP: ip, PrivatePort: p.PrivatePort, PublicPort: p.PublicPort, Type: p.Type})
	}
	return model.ContainerResponse{
		ID:      c.ID,
		Name:    name,
		Image:   c.Image,
		State:   string(c.State),
		Status:  c.Status,
		Created: c.Created,
		Ports:   ports,
		Project: c.Labels["com.docker.compose.project"],
	}
}

func ImageToResponse(i image.Summary) model.ImageResponse {
	tags := i.RepoTags
	if tags == nil {
		tags = []string{}
	}
	return model.ImageResponse{ID: i.ID, Tags: tags, Size: i.Size, Created: i.Created, Containers: i.Containers}
}

func VolumeToResponse(v volume.Volume) model.VolumeResponse {
	return model.VolumeResponse{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, CreatedAt: v.CreatedAt}
}

func NetworkToResponse(n network.Summary) model.NetworkResponse {
	return model.NetworkResponse{ID: n.ID, Name: n.Name, Driver: n.Driver, Scope: n.Scope, Created: n.Created.Unix()}
}

func SystemToResponse(i system.Info) model.SystemResponse {
	return model.SystemResponse{
		Name:              i.Name,
		ServerVersion:     i.ServerVersion,
		OperatingSystem:   i.OperatingSystem,
		KernelVersion:     i.KernelVersion,
		Architecture:      i.Architecture,
		NCPU:              i.NCPU,
		MemTotal:          i.MemTotal,
		Containers:        i.Containers,
		ContainersRunning: i.ContainersRunning,
		ContainersPaused:  i.ContainersPaused,
		ContainersStopped: i.ContainersStopped,
		Images:            i.Images,
	}
}

// StatsToResponse uses the same CPU formula as `docker stats`.
func StatsToResponse(s container.StatsResponse) model.StatsResponse {
	res := model.StatsResponse{MemUsage: s.MemoryStats.Usage, MemLimit: s.MemoryStats.Limit}
	// cgroup v2 reports page cache under inactive_file; docker stats subtracts it.
	if cache, ok := s.MemoryStats.Stats["inactive_file"]; ok && cache < res.MemUsage {
		res.MemUsage -= cache
	}
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	cpus := float64(s.CPUStats.OnlineCPUs)
	if cpus == 0 {
		cpus = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if sysDelta > 0 && cpuDelta > 0 {
		res.CPUPercent = cpuDelta / sysDelta * cpus * 100
	}
	for _, n := range s.Networks {
		res.NetRx += n.RxBytes
		res.NetTx += n.TxBytes
	}
	return res
}

// RedactInspect masks the values of Config.Env in a raw container inspect payload.
// A session already has full Docker access, but the inspect endpoint would otherwise
// hand every other container's secrets (database passwords, API keys) straight to the
// browser, where one XSS or one shared screen leaks them all. The variable names are
// kept because they are what makes the view useful for debugging.
func RedactInspect(raw json.RawMessage) (json.RawMessage, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(doc["Config"], &config); err != nil {
		return raw, nil // no Config object to redact
	}
	var env []string
	if err := json.Unmarshal(config["Env"], &env); err != nil {
		return raw, nil // no Env array to redact
	}
	for i, kv := range env {
		if key, _, ok := strings.Cut(kv, "="); ok {
			env[i] = key + "=" + redacted
		} else {
			env[i] = redacted
		}
	}
	masked, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	config["Env"] = masked
	if doc["Config"], err = json.Marshal(config); err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}

const redacted = "***redacted***"
