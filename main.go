package main

import (
	"errors"
	"math"
	"net/http"

	"github.com/Alexamakans/RemoteControl/failuremessage"
	"github.com/Alexamakans/RemoteControl/keycode"
	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/itchyny/volume-go"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
)

var log = logger.NewScoped("RC")

func main() {
	logger.ClearOutputs()
	logger.AddOutput(logger.LevelDebug, consolepretty.Default)

	engine := gin.New()
	api := engine.Group("/api")
	{
		api.POST("key", onKeyHandler)
	}

	if err := engine.Run("0.0.0.0:6969"); err != nil {
		log.Error().WithError(err).Message("Error during serving.")
	}
}

type onKeyBody struct {
	Key     uint `json:"key"`
	Pressed bool `json:"pressed"`
	Tap     bool `json:"tap"`
}

func onKeyHandler(c *gin.Context) {
	body := onKeyBody{}
	if err := c.BindJSON(&body); err != nil {
		respondJSONError(c, http.StatusInternalServerError, err, failuremessage.JSONBind)
		return
	}

	curVol, err := volume.GetVolume()
	if err != nil {
		respondJSONError(c, http.StatusInternalServerError, err, failuremessage.GetVolume)
		return
	}

	switch body.Key {
	case keycode.VolumeUp:
		if err := volume.SetVolume(int(math.Min(float64(curVol+5), 100))); err != nil {
			respondJSONError(c, http.StatusInternalServerError, err, failuremessage.SetVolume)
			return
		}
	case keycode.VolumeDown:
		if err := volume.SetVolume(int(math.Max(float64(curVol-5), 0))); err != nil {
			respondJSONError(c, http.StatusInternalServerError, err, failuremessage.SetVolume)
			return
		}
	case keycode.Mute:
		if err := volume.Mute(); err != nil {
			respondJSONError(c, http.StatusInternalServerError, err, failuremessage.Mute)
			return
		}
	case keycode.Unmute:
		if err := volume.Unmute(); err != nil {
			respondJSONError(c, http.StatusInternalServerError, err, failuremessage.Unmute)
			return
		}
	case keycode.ToggleMute:
		isMuted, err := volume.GetMuted()
		if err != nil {
			respondJSONError(c, http.StatusInternalServerError, err, failuremessage.GetMuted)
			return
		}
		if isMuted {
			if err := volume.Unmute(); err != nil {
				respondJSONError(c, http.StatusInternalServerError, err, failuremessage.ToggleMute)
				return
			}
		} else {
			if err := volume.Mute(); err != nil {
				respondJSONError(c, http.StatusInternalServerError, err, failuremessage.ToggleMute)
				return
			}
		}
	case keycode.MediaNext:
		respondJSONError(c, http.StatusNotImplemented, errors.New("feature not implemented yet"), failuremessage.NotImplementedYet)
		return
	case keycode.MediaPrevious:
		respondJSONError(c, http.StatusNotImplemented, errors.New("feature not implemented yet"), failuremessage.NotImplementedYet)
		return
	case keycode.MediaPause:
		respondJSONError(c, http.StatusNotImplemented, errors.New("feature not implemented yet"), failuremessage.NotImplementedYet)
		return
	case keycode.MediaPlay:
		respondJSONError(c, http.StatusNotImplemented, errors.New("feature not implemented yet"), failuremessage.NotImplementedYet)
		return
	case keycode.MediaTogglePlay:
		respondJSONError(c, http.StatusNotImplemented, errors.New("feature not implemented yet"), failuremessage.NotImplementedYet)
		return
	default:
		if body.Tap {
			robotgo.KeyTap(string(byte(body.Key)))
		} else {
			if body.Pressed {
				robotgo.KeyDown(string(byte(body.Key)))
			} else {
				robotgo.KeyUp(string(byte(body.Key)))
			}
		}
	}

	// Accepted because not guaranteed that action has completed.
	// It should be successful once completed, though.
	log.Debug().WithUint("key", body.Key).WithBool("pressed", body.Pressed).WithBool("tap", body.Tap).Message("Handled OnKey request successfully.")
	c.JSON(http.StatusAccepted, "Accepted")
}

type errorResponse struct {
	err     string
	message string
}

func respondJSONError(c *gin.Context, code int, err error, message string) {
	errResp := newErrorResponse(err, message)
	log.Warn().WithInt("code", code).WithString("err", errResp.err).WithString("message", errResp.message).Message("Error.")
	c.JSON(code, errResp)
}

func newErrorResponse(err error, message string) errorResponse {
	return errorResponse{
		err:     err.Error(),
		message: message,
	}
}
