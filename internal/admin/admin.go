package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	pb "github.com/iotafs/iotafs-go/internal/protos"
	"github.com/twitchtv/twirp"
)

// ErrVacuumInProgress is returned when a vacuum is already in progress on the server.
var ErrVacuumInProgress = errors.New("vacuum in progress")

// Client performs admin tasks on an IotaFS server.
type Client struct {
	pb.IotaFS
}

// New returns a new admin client.
func New(endpoint string) (*Client, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("unable to parse endpoint: %w", err)
	}
	iclient := pb.NewIotaFSProtobufClient(endpoint, http.DefaultClient)
	return &Client{iclient}, nil
}

// StartVacuum manually initiates a vacuum on the server. It returns immediately. Returns
// ErrVacuumInProgress if a vacuum is currently running on the server.
func (c *Client) StartVacuum(ctx context.Context) (string, error) {
	id, err := c.IotaFS.StartVacuum(ctx, &pb.Empty{})
	if terr, ok := err.(twirp.Error); ok {
		if terr.Code() == twirp.Unavailable {
			return "", ErrVacuumInProgress
		}
	}
	if err != nil {
		return "", err
	}
	return id.Id, nil
}

// Vacuum is returned by VacuumStatus.
type Vacuum struct {
	Status      string
	StartedAt   time.Time
	CompletedAt time.Time
}

// VacuumStatus returns the status of a given vacuum.
func (c *Client) VacuumStatus(ctx context.Context, id string) (Vacuum, error) {
	vacuum, err := c.IotaFS.VacuumStatus(ctx, &pb.VacuumID{Id: id})
	if terr, ok := err.(twirp.Error); ok {
		if terr.Code() == twirp.NotFound {
			return Vacuum{}, fmt.Errorf("vacuum %s not found", id)
		}
	}
	if err != nil {
		return Vacuum{}, err
	}
	startedAt := time.Unix(0, vacuum.StartedAt)
	completedAt := time.Unix(0, vacuum.CompletedAt)
	return Vacuum{vacuum.Status, startedAt, completedAt}, nil
}

type Stats struct {
	NumFiles        uint64
	NumFileVersions uint64
	TotalFilesSize   uint64
	TotalDataSize   uint64
}

// ServerStats returns summary statistics for the server.
func (c *Client) ServerStats(ctx context.Context) (Stats, error) {
	stats, err := c.IotaFS.ServerStats(ctx, &pb.Empty{})
	if err != nil {
		return Stats{}, err
	}
	return Stats{stats.NumFiles, stats.NumFileVersions, stats.TotalFilesSize, stats.TotalDataSize}, nil
}
