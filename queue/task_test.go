package queue

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func MakeTasInstance() TaskInstance {
	return TaskInstance{
		root: NewPath("/temp/test/12345667", afero.NewMemMapFs(), 0777),
		id:   "12345667",
	}
}

func TestTaskInstance_Initialize(t *testing.T) {
	ti := TaskInstance{
		root: NewPath("/temp/test/12345667", afero.NewMemMapFs(), 0777),
		id:   "12345667",
	}
	defer func() {
		err := ti.Remove()
		assert.Nil(t, err)
	}()

	assert.Equal(t, false, ti.IsReady())
	assert.Equal(t, false, ti.IsLocked())
	assert.Equal(t, false, ti.HasError())
	assert.Equal(t, false, ti.Exists())

	// initializing creates the task directory
	err := ti.Initialize()
	assert.Nil(t, err)

	assert.Equal(t, false, ti.IsReady())
	assert.Equal(t, true, ti.IsLocked())
	assert.Equal(t, false, ti.HasError())
	assert.Equal(t, true, ti.Exists())

	// simulate creation of task file
	err = ti.TaskFile().Write([]byte("{}"))
	assert.Nil(t, err)

	err = ti.ReleaseLock()
	assert.Nil(t, err)

	assert.Equal(t, true, ti.IsReady())
	assert.Equal(t, false, ti.IsLocked())
	assert.Equal(t, false, ti.HasError())
	assert.Equal(t, true, ti.Exists())

}

func TestTaskInstance_WriteError(t *testing.T) {
	ti := TaskInstance{
		root: NewPath("/temp/test/12345667", afero.NewMemMapFs(), 0777),
		id:   "12345667",
	}
	defer func() {
		err := ti.Remove()
		assert.Nil(t, err)
	}()
	err := ti.Initialize()
	assert.Nil(t, err)

	err = ti.TaskFile().Write([]byte("{}"))
	assert.Nil(t, err)

	// verify task exists and is locked and has no errors
	assert.Equal(t, true, ti.Exists())
	assert.Equal(t, false, ti.IsReady())
	assert.Equal(t, true, ti.IsLocked())
	assert.Equal(t, false, ti.HasError())

	err = ti.WriteError("test error", "no trace")
	assert.Nil(t, err)

	assert.Equal(t, true, ti.HasError())

	errors, err := ti.GetErrors()
	assert.Nil(t, err)

	assert.Equal(t, 1, errors.Count())

	err = ti.WriteError("test #2", "no trace")
	assert.Nil(t, err)

	errors, err = ti.GetErrors()
	assert.Nil(t, err)

	assert.Equal(t, 2, errors.Count())

}
