package monitoring

import (
	"log/slog"
	"sync"
	"time"

	"utils"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/google/uuid"
)

type netActiveWindowInfo struct {
	windowInfo
	TimeStamp time.Time
	DoNotCopy
}

type DoNotCopy [0]sync.Mutex

type x11DBInterface interface {
	WriteUsage(utils.ScreenTime) error
	UpdateOpertionOnPrefix(dbPrefix string, opsFunc func([]byte) ([]byte, error)) error // these two are here only because I find it useful
	UpdateOperationOnKey(key []byte, opsFunc func([]byte) ([]byte, error)) error        // to manipulate my db, usage in daemon_[os].go (later end of the script).
	GetTaskByUUID(taskID uuid.UUID) (utils.Task, error)
	RemoveTask(id uuid.UUID) error
}

type X11Monitor struct {
	windowChangeCh chan utils.GenericKeyValue[xproto.Window, float64] //windowID and duration
	X11Connection  *xgbutil.XUtil
	Db             x11DBInterface
	logger         *slog.Logger
}

type windowInfo struct {
	WindowID   xproto.Window
	WindowName string
}

type limitWindow struct {
	windowInfo
	taskUUID       uuid.UUID
	timeSofar      float64
	limit          float64
	tenMinToLimit  bool
	fiveMinToLimit bool
}
