// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package persistencetests

import (
	"os"
	"testing"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// HistoryPersistenceSuite contains history persistence tests
	HistoryPersistenceSuite struct {
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

// SetupSuite implementation
func (s *HistoryPersistenceSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// SetupTest implementation
func (s *HistoryPersistenceSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

// TearDownSuite implementation
func (s *HistoryPersistenceSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

func int64Ptr(i int64) *int64 {
	return &(i)
}

// TestAppendHistoryEvents test
func (s *HistoryPersistenceSuite) TestAppendHistoryEvents() {
	domainID := "ff03c29f-fcf1-4aea-893d-1a7ec421e3ec"
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("append-history-events-test"),
		RunId:      common.StringPtr("986fc9cd-4a2d-4964-bf9f-5130116d5851"),
	}

	events1 := &gen.History{Events: []*gen.HistoryEvent{{EventId: int64Ptr(1)}, {EventId: int64Ptr(2)}}}
	err0 := s.AppendHistoryEvents(domainID, workflowExecution, 1, common.EmptyVersion, 1, 1, events1, false)
	s.Nil(err0)

	events2 := &gen.History{Events: []*gen.HistoryEvent{{EventId: int64Ptr(3)}}}
	err1 := s.AppendHistoryEvents(domainID, workflowExecution, 3, common.EmptyVersion, 1, 1, events2, false)
	s.Nil(err1)

	events2New := &gen.History{Events: []*gen.HistoryEvent{{EventId: int64Ptr(4)}}}
	err2 := s.AppendHistoryEvents(domainID, workflowExecution, 3, common.EmptyVersion, 1, 1, events2New, false)
	s.NotNil(err2)
	s.IsType(&p.ConditionFailedError{}, err2)

	// overwrite with higher txnID
	err3 := s.AppendHistoryEvents(domainID, workflowExecution, 3, common.EmptyVersion, 1, 2, events2New, true)
	s.Nil(err3)
}

// TestGetHistoryEvents test
func (s *HistoryPersistenceSuite) TestGetHistoryEvents() {
	domainID := "0fdc53ef-b890-4870-a944-b9b028ac9742"
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("get-history-events-test"),
		RunId:      common.StringPtr("26fa29f6-af41-4b70-9a3b-8b1b35eed82a"),
	}

	batchEvents := newBatchEventForTest([]int64{1, 2}, 1)
	err0 := s.AppendHistoryEvents(domainID, workflowExecution, 1, common.EmptyVersion, 1, 1, batchEvents, false)
	s.Nil(err0)

	// Here the nextEventID is set to 4 to make sure that if NextPageToken is set by persistence, we can get it here.
	// Otherwise the middle layer will clear it. In this way, we can test that if the # of rows got from DB is less than
	// page size, NextPageToken is empty.
	history, token, err1 := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 4, 10, nil)
	s.Nil(err1)
	s.Equal(0, len(token))
	s.Equal(2, len(history.Events))
	s.Equal(int64(1), history.Events[0].GetVersion())

	// We have only one page and the page size is set to one. In this case, persistence may or may not return a nextPageToken.
	// If it does return a token, we need to ensure that if the token returned is used to get history again, no error and history
	// events should be returned.
	history, token, err1 = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 4, 1, nil)
	s.Nil(err1)
	s.Equal(2, len(history.Events))
	if len(token) != 0 {
		history, token, err1 = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 4, 1, token)
		s.Nil(err1)
		s.Equal(0, len(token))
		s.Equal(0, len(history.Events))
	}

	// firstEventID is 2, since there's only one page and nextPageToken is empty,
	// the call should return an error.
	_, _, err2 := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 2, 4, 1, nil)
	s.IsType(&gen.EntityNotExistsError{}, err2)

	// Get history of a workflow that doesn't exist.
	workflowExecution.WorkflowId = common.StringPtr("some-random-id")
	_, _, err2 = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 2, 1, nil)
	s.IsType(&gen.EntityNotExistsError{}, err2)
}

func newBatchEventForTest(eventIDs []int64, version int64) *gen.History {
	var events []*gen.HistoryEvent
	for _, eid := range eventIDs {
		e := &gen.HistoryEvent{EventId: common.Int64Ptr(eid), Version: common.Int64Ptr(version)}
		events = append(events, e)
	}

	return &gen.History{Events: events}
}

// TestGetHistoryEventsCompatibility test
func (s *HistoryPersistenceSuite) TestGetHistoryEventsCompatibility() {
	domainID := "373de9d6-e41e-42d4-bee9-9e06968e4d0d"
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("get-history-events-compatibility-test"),
		RunId:      common.StringPtr(uuid.New()),
	}

	batches := []*gen.History{
		newBatchEventForTest([]int64{1, 2}, 1),
		newBatchEventForTest([]int64{3}, 1),
		newBatchEventForTest([]int64{4, 5, 6}, 1),
		newBatchEventForTest([]int64{6}, 1), // staled batch, should be ignored
		newBatchEventForTest([]int64{7, 8}, 1),
	}

	for i, be := range batches {
		err0 := s.AppendHistoryEvents(domainID, workflowExecution, be.Events[0].GetEventId(), common.EmptyVersion, 1, int64(i), be, false)
		s.Nil(err0)
	}

	// pageSize=3, get 3 batches
	history, token, err := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 8, 3, nil)
	s.Nil(err)
	s.NotNil(token)
	s.Equal(6, len(history.Events))
	for i, e := range history.Events {
		s.Equal(int64(i+1), e.GetEventId())
	}

	// get next page, should ignore staled batch
	history, token, err = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 8, 3, token)
	s.Nil(err)
	s.Nil(token)
	s.Equal(2, len(history.Events))
	s.Equal(int64(7), history.Events[0].GetEventId())
	s.Equal(int64(8), history.Events[1].GetEventId())

	// Start over, but read from middle, should not return error, but the first batch should be ignored by application layer
	token = nil
	history, token, err = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 5, 8, 3, token)
	s.Nil(err)
	s.Nil(token)
	s.Equal(3, len(history.Events))
	s.Equal(int64(6), history.Events[0].GetEventId())
	s.Equal(int64(7), history.Events[1].GetEventId())
	s.Equal(int64(8), history.Events[2].GetEventId())
}

// TestDeleteHistoryEvents test
func (s *HistoryPersistenceSuite) TestDeleteHistoryEvents() {
	domainID := "373de9d6-e41e-42d4-bee9-9e06968e4d0d"
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("delete-history-events-test"),
		RunId:      common.StringPtr("2122fd8d-f583-459e-a2e2-d1fb273a43cb"),
	}

	events := []*gen.History{
		newBatchEventForTest([]int64{1, 2}, 1),
		newBatchEventForTest([]int64{3}, 1),
		newBatchEventForTest([]int64{4, 5}, 1),
		newBatchEventForTest([]int64{5}, 1), // staled batch, should be ignored
		newBatchEventForTest([]int64{6, 7}, 1),
	}
	for i, be := range events {
		err0 := s.AppendHistoryEvents(domainID, workflowExecution, be.Events[0].GetEventId(), common.EmptyVersion, 1, int64(i), be, false)
		s.Nil(err0)
	}

	history, token, err1 := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 8, 11, nil)
	s.Nil(err1)
	s.Nil(token)
	s.Equal(7, len(history.Events))
	for i, e := range history.Events {
		s.Equal(int64(i+1), e.GetEventId())
	}

	err2 := s.DeleteWorkflowExecutionHistory(domainID, workflowExecution)
	s.Nil(err2)

	data1, token1, err3 := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 10, 11, nil)
	s.NotNil(err3)
	s.IsType(&gen.EntityNotExistsError{}, err3)
	s.Nil(token1)
	s.Nil(data1)
}

// TestAppendAndGet test
func (s *HistoryPersistenceSuite) TestAppendAndGet() {
	domainID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("append-and-get-test"),
		RunId:      common.StringPtr(uuid.New()),
	}
	batches := []*gen.History{
		newBatchEventForTest([]int64{1, 2}, 0),
		newBatchEventForTest([]int64{3, 4}, 1),
		newBatchEventForTest([]int64{5, 6}, 2),
		newBatchEventForTest([]int64{7, 8}, 3),
	}

	for i := 0; i < len(batches); i++ {

		events := batches[i].Events
		err0 := s.AppendHistoryEvents(domainID, workflowExecution, events[0].GetEventId(), common.EmptyVersion, 1, int64(i), batches[i], false)
		s.Nil(err0)

		nextEventID := events[len(events)-1].GetEventId()
		history, token, err1 := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, nextEventID, 11, nil)
		s.Nil(err1)
		s.Nil(token)
		s.Equal((i+1)*2, len(history.Events))

		for j, e := range history.Events {
			s.Equal(int64(j+1), e.GetEventId())
		}
	}
}

// TestAppendAndGetByBatch test
func (s *HistoryPersistenceSuite) TestAppendAndGetByBatch() {
	domainID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("append-and-get-test"),
		RunId:      common.StringPtr(uuid.New()),
	}
	batches := []*gen.History{
		newBatchEventForTest([]int64{1, 2}, 0),
		newBatchEventForTest([]int64{3, 4}, 1),
		newBatchEventForTest([]int64{5, 6}, 2),
		newBatchEventForTest([]int64{7, 8}, 3),
	}

	for i := 0; i < len(batches); i++ {

		events := batches[i].Events
		err0 := s.AppendHistoryEvents(domainID, workflowExecution, events[0].GetEventId(), common.EmptyVersion, 1, int64(i), batches[i], false)
		s.Nil(err0)

		nextEventID := events[len(events)-1].GetEventId()

		resp, err1 := s.HistoryMgr.GetWorkflowExecutionHistoryByBatch(&p.GetWorkflowExecutionHistoryRequest{
			DomainID:      domainID,
			Execution:     workflowExecution,
			FirstEventID:  1,
			NextEventID:   nextEventID,
			PageSize:      11,
			NextPageToken: nil,
		})

		s.Nil(err1)
		s.Nil(resp.NextPageToken)

		history := resp.History
		s.Equal((i + 1), len(history))

		for j, h := range history {
			s.Equal(2, len(h.Events))
			s.Equal(int64(j*2+1), h.Events[0].GetEventId())
		}
	}
}

// TestOverwriteAndShadowingHistoryEvents test
func (s *HistoryPersistenceSuite) TestOverwriteAndShadowingHistoryEvents() {
	domainID := "003de9c6-e41e-42d4-bee9-9e06968e4d0d"
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("delete-history-partial-events-test"),
		RunId:      common.StringPtr("2122fd8d-2859-459e-a2e2-d1fb273a43cb"),
	}
	version1 := int64(123)
	version2 := int64(1234)
	var err error

	eventBatches := []*gen.History{
		newBatchEventForTest([]int64{1, 2}, 1),
		newBatchEventForTest([]int64{3}, 1),
		newBatchEventForTest([]int64{4, 5}, 1),
		newBatchEventForTest([]int64{6}, 1),
		newBatchEventForTest([]int64{7}, 1),
		newBatchEventForTest([]int64{8, 9}, 1),
		newBatchEventForTest([]int64{10}, 1),
		newBatchEventForTest([]int64{11, 12}, 1),
		newBatchEventForTest([]int64{13}, 1),
		newBatchEventForTest([]int64{14}, 1),
	}

	for i, be := range eventBatches {
		err = s.AppendHistoryEvents(domainID, workflowExecution, be.Events[0].GetEventId(), version1, 1, int64(i), be, false)
		s.Nil(err)
	}

	history, token, err := s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 15, 25, nil)
	s.Nil(err)
	s.Nil(token)
	s.Equal(14, len(history.Events))
	for i, e := range history.Events {
		s.Equal(int64(i+1), e.GetEventId())
	}

	newEventBatchs := []*gen.History{
		newBatchEventForTest([]int64{8, 9, 10, 11, 12}, 1),
		newBatchEventForTest([]int64{13, 14, 15, 16}, 1),
		newBatchEventForTest([]int64{17, 18}, 1),
		newBatchEventForTest([]int64{19, 20, 21, 22, 23}, 1),
		newBatchEventForTest([]int64{24}, 1),
	}

	for _, be := range newEventBatchs {
		override := false
		for _, oe := range eventBatches {
			if oe.Events[0].GetEventId() == be.Events[0].GetEventId() {
				override = true
				break
			}
		}
		err = s.AppendHistoryEvents(domainID, workflowExecution, be.Events[0].GetEventId(), version2, 1, 999, be, override)
		s.Nil(err)
	}
	historyEvents := []*gen.HistoryEvent{}
	token = nil
	for {
		history, token, err = s.GetWorkflowExecutionHistory(domainID, workflowExecution, 1, 25, 3, token)
		s.Nil(err)
		historyEvents = append(historyEvents, history.Events...)
		if len(token) == 0 {
			break
		}
	}
	s.Empty(token)
	s.Equal(24, len(historyEvents))
	for i, e := range historyEvents {
		s.Equal(int64(i+1), e.GetEventId())
	}
}

// AppendHistoryEvents helper
func (s *HistoryPersistenceSuite) AppendHistoryEvents(domainID string, workflowExecution gen.WorkflowExecution,
	firstEventID, eventBatchVersion int64, rangeID, txID int64, eventsBatch *gen.History, overwrite bool) error {

	_, err := s.HistoryMgr.AppendHistoryEvents(&p.AppendHistoryEventsRequest{
		DomainID:          domainID,
		Execution:         workflowExecution,
		FirstEventID:      firstEventID,
		EventBatchVersion: eventBatchVersion,
		RangeID:           rangeID,
		TransactionID:     txID,
		Events:            eventsBatch.Events,
		Overwrite:         overwrite,
		Encoding:          pickRandomEncoding(),
	})
	return err
}

// GetWorkflowExecutionHistory helper
func (s *HistoryPersistenceSuite) GetWorkflowExecutionHistory(domainID string, workflowExecution gen.WorkflowExecution,
	firstEventID, nextEventID int64, pageSize int, token []byte) (*gen.History, []byte, error) {

	response, err := s.HistoryMgr.GetWorkflowExecutionHistory(&p.GetWorkflowExecutionHistoryRequest{
		DomainID:      domainID,
		Execution:     workflowExecution,
		FirstEventID:  firstEventID,
		NextEventID:   nextEventID,
		PageSize:      pageSize,
		NextPageToken: token,
	})

	if err != nil {
		return nil, nil, err
	}

	return response.History, response.NextPageToken, nil
}

// DeleteWorkflowExecutionHistory helper
func (s *HistoryPersistenceSuite) DeleteWorkflowExecutionHistory(domainID string,
	workflowExecution gen.WorkflowExecution) error {

	return s.HistoryMgr.DeleteWorkflowExecutionHistory(&p.DeleteWorkflowExecutionHistoryRequest{
		DomainID:  domainID,
		Execution: workflowExecution,
	})
}
