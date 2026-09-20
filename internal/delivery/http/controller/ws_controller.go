package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/idnatiya/netradock/internal/delivery/http/middleware"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/sirupsen/logrus"
)

// execStartTimeout bounds how long a shell session may sit idle before sending its start message.
const execStartTimeout = 10 * time.Second

// sessionCheckInterval is how often a live stream re-checks the session that opened it.
const sessionCheckInterval = 30 * time.Second

type WSController struct {
	Log     *logrus.Logger
	UseCase *usecase.ContainerUseCase
	Auth    *usecase.AuthUseCase
}

func NewWSController(log *logrus.Logger, useCase *usecase.ContainerUseCase, auth *usecase.AuthUseCase) *WSController {
	return &WSController{Log: log, UseCase: useCase, Auth: auth}
}

// guardSession tears the connection down once the cookie that opened it stops verifying.
// The auth middleware only runs at upgrade time, so without this a shell or log stream
// would outlive both the session expiry and an explicit logout.
//
// It tears down by expiring the read deadline rather than by calling Close or writing a
// close frame: fasthttp makes Close a no-op on hijacked connections, and another goroutine
// may be mid-write. Failing the pending read unwinds each handler down its normal exit path,
// which is what actually drops the TCP connection.
func (c *WSController) guardSession(conn *websocket.Conn, ctx context.Context, cancel context.CancelFunc) {
	token := conn.Cookies(middleware.SessionCookie)
	go func() {
		ticker := time.NewTicker(sessionCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !c.Auth.Verify(token) {
					c.Log.Debug("websocket closed: session no longer valid")
					cancel()
					_ = conn.SetReadDeadline(time.Now())
					return
				}
			}
		}
	}()
}

// wsWriter sends every Write as one binary frame.
type wsWriter struct{ conn *websocket.Conn }

func (w wsWriter) Write(p []byte) (int, error) {
	if err := w.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// untilClosed returns a context that is cancelled once the client disconnects.
// It owns reading from conn, so only use it when the client sends nothing.
func untilClosed(conn *websocket.Conn) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	return ctx, cancel
}

func (c *WSController) closeWithError(conn *websocket.Conn, err error) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
		return
	}
	c.Log.WithError(err).Debug("websocket stream ended")
	msg := err.Error()
	if len(msg) > 120 {
		msg = msg[:120]
	}
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, msg))
}

func (c *WSController) Logs(conn *websocket.Conn) {
	ctx, cancel := untilClosed(conn)
	defer cancel()
	c.guardSession(conn, ctx, cancel)
	tail := 200
	if n, err := strconv.ParseUint(conn.Query("tail"), 10, 32); err == nil && n <= 10000 {
		tail = int(n)
	}
	err := c.UseCase.StreamLogs(ctx, conn.Params("id"), tail, wsWriter{conn})
	c.closeWithError(conn, err)
}

func (c *WSController) Stats(conn *websocket.Conn) {
	ctx, cancel := untilClosed(conn)
	defer cancel()
	c.guardSession(conn, ctx, cancel)
	err := c.UseCase.StreamStats(ctx, conn.Params("id"), func(s model.StatsResponse) error {
		return conn.WriteJSON(s)
	})
	c.closeWithError(conn, err)
}

func (c *WSController) AllStats(conn *websocket.Conn) {
	ctx, cancel := untilClosed(conn)
	defer cancel()
	c.guardSession(conn, ctx, cancel)
	err := c.UseCase.StreamAllStats(ctx, func(s map[string]model.StatsResponse) error {
		return conn.WriteJSON(s)
	})
	c.closeWithError(conn, err)
}

type execControl struct {
	Type string   `json:"type"`
	Cmd  []string `json:"cmd"`
	Cols uint     `json:"cols"`
	Rows uint     `json:"rows"`
}

// readStart waits for the opening {"type":"start","cmd":[...]} control message.
// The command travels in the body rather than the query string so it stays out of
// reverse proxy access logs and browser history.
func readStart(conn *websocket.Conn) ([]string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(execStartTimeout)); err != nil {
		return nil, err
	}
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()
	_, data, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ctl execControl
	if err := json.Unmarshal(data, &ctl); err != nil || ctl.Type != "start" {
		return nil, errors.New("expected a start control message")
	}
	if len(ctl.Cmd) == 0 {
		return []string{"/bin/sh"}, nil
	}
	return ctl.Cmd, nil
}

// Exec bridges a TTY exec session: binary frames are stdin, text frames are JSON control messages.
func (c *WSController) Exec(conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.guardSession(conn, ctx, cancel)

	cmd, err := readStart(conn)
	if err != nil {
		c.closeWithError(conn, err)
		return
	}
	execID, hijack, err := c.UseCase.Exec(ctx, conn.Params("id"), cmd)
	if err != nil {
		c.closeWithError(conn, err)
		return
	}
	defer hijack.Conn.Close()
	c.Log.WithFields(logrus.Fields{"container": conn.Params("id"), "cmd": cmd}).Info("exec started")

	go func() {
		defer cancel()
		defer hijack.Conn.Close()
		for {
			kind, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if kind == websocket.BinaryMessage {
				if _, err := hijack.Conn.Write(data); err != nil {
					return
				}
				continue
			}
			var ctl execControl
			if json.Unmarshal(data, &ctl) == nil && ctl.Type == "resize" && ctl.Cols > 0 && ctl.Rows > 0 {
				_ = c.UseCase.ExecResize(ctx, execID, ctl.Cols, ctl.Rows)
			}
		}
	}()

	_, err = io.Copy(wsWriter{conn}, hijack.Reader)
	if ctx.Err() == nil {
		c.closeWithError(conn, err)
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "exited"))
	}
}
