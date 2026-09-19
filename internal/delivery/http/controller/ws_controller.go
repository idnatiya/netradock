package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/sirupsen/logrus"
)

type WSController struct {
	Log     *logrus.Logger
	UseCase *usecase.ContainerUseCase
}

func NewWSController(log *logrus.Logger, useCase *usecase.ContainerUseCase) *WSController {
	return &WSController{Log: log, UseCase: useCase}
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
	err := c.UseCase.StreamStats(ctx, conn.Params("id"), func(s model.StatsResponse) error {
		return conn.WriteJSON(s)
	})
	c.closeWithError(conn, err)
}

type execControl struct {
	Type string `json:"type"`
	Cols uint   `json:"cols"`
	Rows uint   `json:"rows"`
}

// Exec bridges a TTY exec session: binary frames are stdin, text frames are JSON control messages.
func (c *WSController) Exec(conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := strings.Fields(conn.Query("cmd", "/bin/sh"))
	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
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
