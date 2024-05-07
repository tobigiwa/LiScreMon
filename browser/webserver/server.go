package webserver

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
)

type Message struct {
	Endpoint   string         `json:"endpoint"`
	SliceData  []KeyValuePair `json:"sliceData"`
	StringData string         `json:"stringData"`
	IntData    int            `json:"intData"`
	IsError    bool           `json:"isError"`
	Error      error          `json:"error"`
}

type KeyValuePair struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

func (m *Message) encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Message) decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(m); err != nil {
		return err
	}
	return nil
}

func (m *Message) decodeToJson() ([]byte, error) {
	return json.Marshal(m)
}

type App struct {
	logger     *slog.Logger
	daemonConn net.Conn
}

func NewApp(logger *slog.Logger) (*App, error) {
	daemonConn, err := listenToDaemonService()
	if err != nil {
		return nil, err
	}

	return &App{
		logger:     logger,
		daemonConn: daemonConn,
	}, nil
}

func listenToDaemonService() (net.Conn, error) {
	var (
		unix     = "unix"
		homeDir  string
		err      error
		unixAddr *net.UnixAddr
	)
	if homeDir, err = os.UserHomeDir(); err != nil {
		return nil, err
	}
	socketDir := homeDir + "/liScreMon/socket/daemon.sock"

	if unixAddr, err = net.ResolveUnixAddr(unix, socketDir); err != nil {
		return nil, err
	}

	conn, err := net.DialUnix(unix, nil, unixAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (a *App) CheckDaemonService() error {
	msg := Message{
		Endpoint:   "startConnection",
		StringData: "I wish this project prospered.",
	}
	bytes, err := msg.encode()
	if err != nil {
		return err
	}
	if _, err = a.daemonConn.Write(bytes); err != nil {
		return err
	}

	buf := make([]byte, 512)
	if _, err := a.daemonConn.Read(buf); err != nil {
		return err
	}
	if err = msg.decode(buf); err != nil {
		return err
	}
	fmt.Printf("%+v\n\n", msg)
	return nil
}