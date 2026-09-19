package converter

import (
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
