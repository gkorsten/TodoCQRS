package eventbus

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	TestEvent Event = "testEvent"
	TestEvent1 Event = "UserCreatedEvent"
	TestEvent2 Event = "OrderPlacedEvent"
)

func HandlerT(ev Event) []byte {
	time.Sleep(10 * time.Millisecond)
	 return []byte{}
}

func TestRegisterUnregister (t *testing.T) {
	x:=NewEventBus()

	//register two events 
	id := x.RegisterEvent(TestEvent,HandlerT) 
	id2 := x.RegisterEvent(TestEvent2,HandlerT) 
	assert.Equal(t,len(x.clients),2,fmt.Sprintf("register failed have %v clients should have 2",len(x.clients)))

	//unresgister 1 of the events
	x.Unregister(TestEvent,id)
	assert.Equal(t,len(x.clients),1,fmt.Sprintf("unregister failed have %v clients should have 1",len(x.clients)))

	//test that it errrors if a event isn't registerd anymore
	err:= x.Unregister(TestEvent,id)
	assert.Error(t, err, "Unregiter an event that doesn't exist should produce an error")
	assert.Contains(t, err.Error(), "no such event", "Error message should indicate no event")

	//test that it errors if you send the wrong id was sent
	err= x.Unregister(TestEvent2,id)
	assert.Error(t, err, "Unregiter an event that doesn't exist should produce an error")
	assert.Contains(t, err.Error(), "no such handler or id found", "Error message should indicate no event or id")

	//test that we can unregister event 2
	err= x.Unregister(TestEvent2,id2)
	assert.NoError(t,err,"Should not get an error, as Testevent2 and id2 should match")
	assert.Equal(t,len(x.clients),0,fmt.Sprintf("unregister failed have %v clients should have 0",len(x.clients)))
}

func TestEventBus_Publish_Success(t *testing.T) {
	// 1. Setup
	bus := NewEventBus()
	defer bus.Stop()

	// 2. Define event and counter
	var counter atomic.Int32
	numHandlers := 5

	// 3. Register multiple handlers
	for i := 0; i < numHandlers; i++ {
		// Define the handler function
		handler := func(e Event)[]byte {
			// Simulate work or processing time
			time.Sleep(10 * time.Millisecond)
			
			// Concurrently increment the counter
			counter.Add(1)
			 return []byte{}
		}
		
		// Register the handler
		bus.RegisterEvent(TestEvent1, handler)
	}

	// Wait briefly to ensure the handlers are registered before publishing
	time.Sleep(10 * time.Millisecond)

	// 4. Publish the event
	err := bus.Publish(TestEvent1)

	// Assert that publishing was successful
	assert.NoError(t, err, "Publishing should not return an error")

	// there is a race condition here - as the test can exit before the Run event has incremented the WaitGroup
    time.Sleep(10 * time.Millisecond) 
	bus.wait()
	// 6. Assert that all handlers ran
	finalCount := counter.Load()
	assert.Equal(t, int32(numHandlers), finalCount, 
		fmt.Sprintf("Expected %d handlers to run, but got %d", numHandlers, finalCount))
	
}

func TestEventBus_Publish_NoSubscriber(t *testing.T) {
	bus := NewEventBus()
	defer bus.Stop()

	// Publish an event that has no registered subscribers
	err := bus.Publish(TestEvent2)

	// Assert that an error is returned because the event is not subscribed
	assert.Error(t, err, "Publishing an unsubscribed event should return an error")
	assert.Contains(t, err.Error(), "no such event or subscriptions", "Error message should indicate no subscription")
}

func TestEventBus_Publish_ClosedBus(t *testing.T) {
	bus := NewEventBus()
	
	// Stop the bus immediately
	bus.Stop()

	// Attempt to publish after stopping
	err := bus.Publish(TestEvent1)

	// Assert that an error is returned because the bus is closed
	assert.Error(t, err, "Publishing on a closed bus should return an error")
	assert.Contains(t, err.Error(), "event bus is down", "Error message should indicate bus is down")
}