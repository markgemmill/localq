package queue

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MasterQSuite struct {
	suite.Suite
	fs     afero.Fs
	master *MasterQ
}

func (suite *MasterQSuite) SetupSuite() {
	suite.fs = afero.NewMemMapFs()

	master, _ := New("/local", suite.fs, 0777)

	err := master.Register(&ConcreteTask{}, "concrete")
	assert.Nil(suite.T(), err)

	suite.master = master
}

func (suite *MasterQSuite) TeardownSuite() {
	err := suite.fs.Remove("/")
	assert.Nil(suite.T(), err)
}

func (suite *MasterQSuite) TestMasterQ_New() {
	masterA := suite.master

	masterB, err := New("/local2", suite.fs, 0777)
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), 2, len(globalQ))

	masterC, err := New("/local", suite.fs, 0777)
	assert.Nil(suite.T(), err)

	assert.NotSame(suite.T(), masterA, masterB)
	assert.Same(suite.T(), masterA, masterC)
}

func (suite *MasterQSuite) TestMasterQ_Register() {
	assert.True(suite.T(), suite.master.Has("concrete"))
	assert.False(suite.T(), suite.master.Has("stone"))

	tq, err := suite.master.Get("concrete")
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), "/local/concrete", tq.root.String())
	assert.True(suite.T(), tq.root.Exists())

}

func (suite *MasterQSuite) TestMasterQ_Enqueue() {
	master := suite.master

	tq := master.Enqueue("concrete")

	assert.Equal(suite.T(), "concrete", tq.name)
	assert.True(suite.T(), tq.root.Exists())

	InvalidEnqueue := func() {
		master.Enqueue("stone")
	}

	assert.PanicsWithError(suite.T(), "task 'stone' is not registered", InvalidEnqueue)
}

func TestMasterQSuite(t *testing.T) {
	suite.Run(t, new(MasterQSuite))
}
