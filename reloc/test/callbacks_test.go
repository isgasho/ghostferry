package test

import (
	"encoding/json"
	"net/http"
	"testing"

	rtesthelpers "github.com/Shopify/ghostferry/reloc/testhelpers"
	"github.com/Shopify/ghostferry/testhelpers"

	"github.com/stretchr/testify/suite"
)

type CallbacksTestSuite struct {
	*rtesthelpers.RelocUnitTestSuite

	errHandler testhelpers.ErrorHandler
}

func (t *CallbacksTestSuite) SetupTest() {
	t.RelocUnitTestSuite.SetupTest()

	t.Ferry.Ferry.ErrorHandler = &t.errHandler

	err := t.Ferry.Start()
	t.Require().Nil(err)
}

func (t *CallbacksTestSuite) TearDownTest() {
	t.RelocUnitTestSuite.TearDownTest()
}

func (t *CallbacksTestSuite) TestFailsRunOnUnlockError() {
	callbackReceived := false
	t.CutoverUnlock = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callbackReceived = true
		w.WriteHeader(http.StatusInternalServerError)
	})

	t.Ferry.Run()

	t.Require().True(callbackReceived)

	t.Require().NotNil(t.errHandler.LastError)
	t.Require().Equal("callback returned 500 Internal Server Error", t.errHandler.LastError.Error())
}

func (t *CallbacksTestSuite) TestFailsRunOnLockError() {
	callbackReceived := false
	t.CutoverLock = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callbackReceived = true
		w.WriteHeader(http.StatusInternalServerError)
	})

	t.Ferry.Run()

	t.Require().True(callbackReceived)

	t.Require().NotNil(t.errHandler.LastError)
	t.Require().Equal("callback returned 500 Internal Server Error", t.errHandler.LastError.Error())
}

func (t *CallbacksTestSuite) TestPostsCallbacks() {
	lockReceived := false
	t.CutoverLock = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lockReceived = true
		resp := t.requestMap(r)
		t.Require().Equal("test_lock", resp["Payload"])
	})

	unlockReceived := false
	t.CutoverUnlock = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unlockReceived = true
		resp := t.requestMap(r)
		t.Require().Equal("test_unlock", resp["Payload"])
	})

	t.Ferry.Run()

	t.Require().True(lockReceived)
	t.Require().True(unlockReceived)

	t.AssertTenantCopied()
}

func (t *CallbacksTestSuite) requestMap(r *http.Request) map[string]string {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)

	var res map[string]string
	err := decoder.Decode(&res)
	t.Require().Nil(err)

	return res
}

func TestCallbacksTestSuite(t *testing.T) {
	suite.Run(t, &CallbacksTestSuite{RelocUnitTestSuite: &rtesthelpers.RelocUnitTestSuite{}})
}