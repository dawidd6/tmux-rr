package tmux_test

import (
	"testing"

	"github.com/dawidd6/tmux-rr/pkg/tmux"
	"github.com/stretchr/testify/assert"
)

func TestSerializeFormat(t *testing.T) {
	expected := &tmux.Session{
		Name: "#{session_name}",
	}
	actual := tmux.SerializeFormat[tmux.Session]()
	assert.EqualValues(t, expected, actual)
}

func TestSerializeFormatJSON(t *testing.T) {
	expected := `{"name":"#{session_name}"}`
	actual, err := tmux.SerializeFormatJSON[tmux.Session]()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestTest(t *testing.T) {
	sessions, err := tmux.ListObjects[tmux.Session]()
	assert.NoError(t, err)
	for _, session := range sessions {
		t.Logf("%+v\n", session)
	}
}
